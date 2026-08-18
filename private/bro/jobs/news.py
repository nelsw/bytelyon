import asyncio
import logging
from dataclasses import dataclass
from datetime import datetime as dt
from datetime import tzinfo
from urllib.parse import quote

import aiohttp
from aiohttp import ClientSession
from playwright.async_api import BrowserContext, Error
from pytz import timezone

from jobs.job import Job
from models.bot import Bot
from services.redis import publish_news
from utils.fetch import fetch_xml

RFC_1123: str = "%a, %d %b %Y %H:%M:%S %Z"
UTC: tzinfo = timezone("UTC")

logger = logging.getLogger(__name__)


@dataclass
class Article:
    url: str
    published_at: dt
    title: str


class NewsJob(Job):
    def __init__(
        self,
        bot: Bot,
        max_concurrency: int = 5,
        max_retries: int = 3,
        retry_backoff: float = 3.0,
    ):
        super().__init__(
            bot,
            max_concurrency,
            max_retries,
            retry_backoff,
        )
        self.bot = bot
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
                logger.info(f"[+] Scraping: {a.url}{suffix}")
                page = await ctx.new_page()
                try:
                    await page.goto(a.url, wait_until="domcontentloaded", timeout=5000)
                    await publish_news(
                        bot_id=self.bot.id,
                        published_at=a.published_at,
                        title=a.title,
                        page=page,
                    )
                    logger.info(f"[+] Appended Article: {page.url}")
                    return
                except Error:
                    if attempt <= self.max_retries:
                        delay = self.retry_backoff * attempt
                        print(
                            f"[!] {page.url} failed ({last_error}); retrying in {delay:.0f}s"
                        )
                        await asyncio.sleep(delay)
                    else:
                        print(
                            f"[-] Giving up on {page.url} after {self.max_retries + 1} attempts"
                        )
                finally:
                    await page.close()

    async def task(self, context: BrowserContext):
        while True:
            a: Article = await self.queue.get()
            if a.url == "chrome-error://chromewebdata/":
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
                at: dt = dt.strptime(
                    e.findtext("pubDate", default=""), RFC_1123
                ).astimezone(tz=UTC)
                if at < self.bot.last_ran_at:
                    await self.queue.put(
                        Article(
                            url=e.findtext("link", default=""),
                            published_at=at,
                            title=e.findtext("title", default=""),
                        )
                    )

    async def pre_process(self) -> None:
        urls = [
            f"https://www.bing.com/news/search?format=rss&q={quote(self.bot.query)}",
            f"https://news.google.com/rss/search?q={quote(self.bot.query)}&hl=en-US&gl=US&ceid=US:en",
        ]
        async with aiohttp.ClientSession() as session:
            await asyncio.gather(*[self.fetch(session, url) for url in urls])
