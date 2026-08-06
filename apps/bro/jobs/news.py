import argparse
import asyncio
import datetime
from datetime import datetime as dt
from urllib.parse import quote

import aiohttp
from aiohttp import ClientSession
from playwright.async_api import BrowserContext, Error, async_playwright
from seleniumbase.undetected import cdp_driver

from models.article import Article, from_element
from utils.fetch import fetch_xml


class NewsBot:
    def __init__(
        self,
        bot_id: int,
        query: str,
        since: datetime.datetime,
        max_concurrency: int,
        max_pages: int,
        max_retries: int = 3,
        retry_backoff: float = 3.0,
    ):
        self.bot_id = bot_id
        self.query = query
        self.since = since
        self.articles: list[Article] = []
        self.max_concurrency = max_concurrency
        self.bounder = asyncio.Semaphore(max_concurrency)
        self.queue: asyncio.Queue = asyncio.Queue()
        self.visited_urls: set = set()
        self.visited_lock = asyncio.Lock()
        self.pages_crawled = 0
        self.max_pages = max_pages
        self.max_retries = max_retries
        self.retry_backoff = retry_backoff

    async def scrape_page(self, ctx: BrowserContext, a: Article) -> None:
        last_error = None
        for attempt in range(1, self.max_retries + 2):
            async with self.bounder:
                suffix = (
                    f" (attempt {attempt}/{self.max_retries + 1})"
                    if attempt > 1
                    else ""
                )
                print(f"[+] Scraping: {a.url}{suffix}")
                page = await ctx.new_page()
                try:
                    await page.goto(a.url, wait_until="domcontentloaded", timeout=10000)
                    await page.wait_for_selector("body", timeout=5000)
                    if a.source == "Google News":
                        await asyncio.sleep(1)
                    self.articles.append(
                        a.with_data(url=page.url, html=await page.content())
                    )
                    print(f"[+] Appended Article: {page.url}")
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

    async def worker(self, ctx: BrowserContext):
        """Pulls URLs off the queue and processes them until the queue drains.

        Once max_pages is hit, remaining queued URLs are drained (marked done,
        not visited) rather than left in place, so queue.join() can still
        complete instead of waiting forever on entries nobody will process.
        """
        while True:
            a: Article = await self.queue.get()
            try:
                async with self.visited_lock:
                    self.visited_urls.add(a.url)
                    self.pages_crawled += 1
                await self.scrape_page(ctx, a)
            finally:
                self.queue.task_done()

    async def fetch(self, session: ClientSession, url: str) -> None:
        xml = await fetch_xml(session, url)
        if xml is not None:
            for e in xml.findall(".//item"):
                a = from_element(element=e, after=self.since)
                if a is not None:
                    await self.queue.put(a)

    async def run(self) -> None:
        async with aiohttp.ClientSession() as session:
            await asyncio.gather(
                *[
                    self.fetch(session, url)
                    for url in [
                        f"https://www.bing.com/news/search?format=rss&q={quote(self.query)}",
                        f"https://news.google.com/rss/search?q={quote(self.query)}&hl=en-US&gl=US&ceid=US:en",
                    ]
                ]
            )

        driver = await cdp_driver.start_async()
        endpoint_url = driver.get_endpoint_url()
        async with async_playwright() as p:
            browser = await p.chromium.connect_over_cdp(endpoint_url)
            context = browser.contexts[0]
            workers = [
                asyncio.create_task(self.worker(context))
                for _ in range(self.max_concurrency)
            ]
            await self.queue.join()
            for worker_task in workers:
                worker_task.cancel()
            await asyncio.gather(*workers, return_exceptions=True)
            await browser.close()


def main():
    parser = argparse.ArgumentParser(
        description="Get XML from Google and Bing News RSS, visit items and save page source."
    )
    parser.add_argument("bot_id", type=int, help="Unique identifier for the bot")
    parser.add_argument("query", type=str, help="The search query")
    parser.add_argument(
        "since", type=str, help="The oldest date for including articles"
    )
    parser.add_argument(
        "--max-concurrency", type=int, default=5, help="Concurrent pages (default: 5)"
    )
    parser.add_argument(
        "--max-pages",
        type=int,
        default=200,
        help="Hard stop on pages crawled (default: 200)",
    )
    parser.add_argument(
        "--out-dir", default="output", help="Output directory (default: output)"
    )
    parser.add_argument(
        "--max-retries",
        type=int,
        default=5,
        help="Retries per failed page, beyond the first attempt (default: 5)",
    )
    parser.add_argument(
        "--retry-backoff",
        type=float,
        default=2.0,
        help="Base seconds to wait before a retry, multiplied by attempt number (default: 2.0)",
    )
    args = parser.parse_args()

    asyncio.run(
        NewsBot(
            bot_id=args.bot_id,
            query=args.query,
            since=dt.fromisoformat(args.since),
            max_concurrency=args.max_concurrency,
            max_pages=args.max_pages,
            max_retries=args.max_retries,
            retry_backoff=args.retry_backoff,
        ).run()
    )


if __name__ == "__main__":
    main()
