from models.bot import Bot
from dataclasses import asdict

import httpx

from services.http import put_article, put_bot, del_bot
import logging
from typing import Optional
import asyncio
import datetime
from urllib.parse import quote

import aiohttp
from aiohttp import ClientSession
from playwright.async_api import BrowserContext, Error

from jobs.job import Job
from models.article import Article, from_element, Articles
from utils.fetch import fetch_xml

logger = logging.getLogger(__name__)

class NewsJob(Job):
    def __init__(
            self,
            bot: Bot,
            max_concurrency: int = 5,
            max_retries: int = 3,
            retry_backoff: float | int = 3.0,
    ):
        super().__init__(
            bot.headless,
            max_concurrency,
            max_retries,
            retry_backoff,
        )
        self.bot_id = bot.id
        self.topic = bot.query
        self.since = bot.last_ran_at
        self.articles: Articles = []
        self.visited_urls: set = set()


    async def scrape_page(self, ctx: BrowserContext, a: Article) -> None:
        last_error = None
        for attempt in range(1, self.max_retries + 2):
            async with self.bounder:
                suffix = (
                    f" (attempt {attempt}/{self.max_retries + 1})"
                    if attempt > 1
                    else ""
                )
                logging.info(f"[+] Scraping: {a.url}{suffix}")
                page = await ctx.new_page()
                try:
                    await page.goto(a.url, wait_until="domcontentloaded", timeout=5000)
                    self.articles.append(await a.with_data(page))
                    logging.info(f"[+] Appended Article: {page.url}")
                    return
                except Error as e:
                    last_error = e
                finally:
                    await page.close()

            if attempt <= self.max_retries:
                delay = self.retry_backoff * attempt
                print(f"[!] {page.url} failed ({last_error}); retrying in {delay:.0f}s")
                await asyncio.sleep(delay)

        print(f"[-] Giving up on {page.url} after {self.max_retries + 1} attempts")
        return

    async def task(self, context: BrowserContext):
        while True:
            a: Article = await self.queue.get()
            if a.url == 'chrome-error://chromewebdata/':
                continue
            try:
                async with self.work_lock:
                    if a.url in self.visited_urls:
                        continue
                    self.visited_urls.add(a.url)
                await self.scrape_page(context, a)
            finally:
                self.queue.task_done()

    async def fetch(self, session: ClientSession, url: str) -> None:
        xml = await fetch_xml(session, url)
        if xml is not None:
            for e in xml.findall(".//item"):
                a = from_element(element=e, after=self.since)
                if a is not None:
                    await self.queue.put(a)

    async def pre_process(self) -> None:
        urls = [
            f"https://www.bing.com/news/search?format=rss&q={quote(self.topic)}",
            f"https://news.google.com/rss/search?q={quote(self.topic)}&hl=en-US&gl=US&ceid=US:en",
        ]
        async with aiohttp.ClientSession() as session:
            await asyncio.gather(*[self.fetch(session, url) for url in urls])

    async def post_process(self) -> None:
        async with httpx.AsyncClient() as c:
            tasks = [put_article(c, self.bot_id, a) for a in self.articles]
            tasks.append(put_bot(c, self.bot_id))
            tasks.append(del_bot(c, self.bot_id))
            await asyncio.gather(*tasks)
