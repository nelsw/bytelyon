import argparse
import asyncio
import datetime
import os
from datetime import datetime as dt
from urllib.parse import quote, urlparse
from xml.etree import ElementTree

import aiohttp
from aiohttp import ClientSession
from playwright.async_api import (
    BrowserContext,
    Error,
    async_playwright,
)
from pytz import timezone
from seleniumbase.undetected import cdp_driver

from models.doc import Doc

RFC_1123 = "%a, %d %b %Y %H:%M:%S %Z"


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
        self.articles: list[dict] = []
        self.max_concurrency = max_concurrency
        self.bounder = asyncio.Semaphore(max_concurrency)
        self.queue: asyncio.Queue = asyncio.Queue()
        self.visited_urls: set = set()
        self.visited_lock = asyncio.Lock()
        self.pages_crawled = 0
        self.max_pages = max_pages
        self.max_retries = max_retries
        self.retry_backoff = retry_backoff
        self.output_dir = f"output/{self.bot_id}"
        os.makedirs(self.output_dir, exist_ok=True)
        print(f"[ ] NewsBot: id={self.bot_id} query='{self.query}' since={self.since}")

    def append_article(self, item: dict, url: str, html: str) -> None:
        d = Doc(html)
        item["url"] = url
        if item["title"] == "":
            item["title"] = d.title()
        item["description"] = d.description()
        item["publisher"] = d.source()
        if item["publisher"] == "":
            item["publisher"] = urlparse(url).netloc.removeprefix("www.")
        item["body"] = d.body()
        item["description"] = d.description()
        item["img_url"] = d.img_url()
        item["img_alt"] = d.img_alt()
        item["keywords"] = d.keywords()
        self.articles.append(item)

    async def scrape_page(self, context: BrowserContext, item: dict) -> None:
        last_error = None
        url = item["url"]
        for attempt in range(1, self.max_retries + 2):
            async with self.bounder:
                suffix = (
                    f" (attempt {attempt}/{self.max_retries + 1})"
                    if attempt > 1
                    else ""
                )
                print(f"[+] Scraping: {url}{suffix}")
                page = await context.new_page()
                try:
                    await page.goto(url, wait_until="domcontentloaded", timeout=10000)
                    await page.wait_for_selector("body", timeout=5000)
                    if item["source"] == "Google News":
                        await asyncio.sleep(1)
                    url = page.url
                    content = await page.content()
                    self.append_article(item, url, content)
                    print(f"[+] Appended {item['source']} Article: {url}")
                    return
                except Error as e:
                    last_error = e
                finally:
                    await page.close()

            if attempt <= self.max_retries:
                delay = self.retry_backoff * attempt
                print(f"[!] {url} failed ({last_error}); retrying in {delay:.0f}s")
                await asyncio.sleep(delay)

        print(f"[-] Giving up on {url} after {self.max_retries + 1} attempts")
        return

    async def worker(self, context: BrowserContext):
        """Pulls URLs off the queue and processes them until the queue drains.

        Once max_pages is hit, remaining queued URLs are drained (marked done,
        not visited) rather than left in place, so queue.join() can still
        complete instead of waiting forever on entries nobody will process.
        """
        while True:
            item = await self.queue.get()
            try:
                async with self.visited_lock:
                    self.visited_urls.add(item["url"])
                    self.pages_crawled += 1

                await self.scrape_page(context, item)

            finally:
                self.queue.task_done()

    async def fetch_url(self, session: ClientSession, url: str) -> None:
        # Sends an asynchronous GET request
        async with session.get(url) as response:
            if response.status >= 300:
                print(f"[!] Error fetching {url}")
                return

            source: str = "Bing News"
            if url.startswith("https://news.google"):
                source = "Google News"

            text = await response.text(encoding="utf-8")
            for item in ElementTree.fromstring(text=text).findall(".//item"):
                pubdate = item.findtext("pubDate", default="")
                if (
                    dt.strptime(pubdate, RFC_1123).astimezone(timezone("UTC"))
                    > self.since
                ):
                    await self.queue.put(
                        {
                            "source": source,
                            "url": item.findtext("link"),
                            "title": item.findtext("title"),
                            "published_at": pubdate,
                        }
                    )

    async def run(self) -> None:
        urls = [
            f"https://www.bing.com/news/search?format=rss&q={quote(self.query)}",
            f"https://news.google.com/rss/search?q={quote(self.query)}&hl=en-US&gl=US&ceid=US:en",
        ]
        async with aiohttp.ClientSession() as session:
            tasks = [self.fetch_url(session, url) for url in urls]
            await asyncio.gather(*tasks)

        driver = await cdp_driver.start_async()
        try:
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
        finally:
            driver.stop()


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

    bot = NewsBot(
        bot_id=args.bot_id,
        query=args.query,
        since=dt.fromisoformat(args.since),
        max_concurrency=args.max_concurrency,
        max_pages=args.max_pages,
        max_retries=args.max_retries,
        retry_backoff=args.retry_backoff,
    )

    asyncio.run(bot.run())


if __name__ == "__main__":
    main()
