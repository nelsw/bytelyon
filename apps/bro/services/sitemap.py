import argparse
import asyncio
import json
import os
import uuid
from urllib.parse import urlparse

import aiofiles
from doc import Doc
from playwright.async_api import (  # pyright: ignore[reportMissingImports]
    BrowserContext,
    Error,
    Page,
    async_playwright,
)
from seleniumbase import cdp_driver  # pyright: ignore[reportMissingImports]


class SitemapBot:
    """Crawls a domain-breadth-first, fetching multiple pages concurrently.

    Chrome is launched via SeleniumBase's undetected cdp_driver (to reduce
    bot-detection) and driven with playwright.async_api for page automation.
    """

    def __init__(
        self,
        bot_id: int,
        domain: str,
        max_concurrency: int,
        max_pages: int,
        max_retries: int = 3,
        retry_backoff: float = 3.0,
    ):
        self.bot_id = bot_id
        self.domain = domain
        self.max_concurrency = max_concurrency
        self.bounder = asyncio.Semaphore(max_concurrency)
        self.queue: asyncio.Queue = asyncio.Queue()
        self.visited_urls: set = set()
        self.pages = []
        self.visited_lock = asyncio.Lock()
        self.pages_crawled = 0
        self.max_pages = max_pages
        self.max_retries = max_retries
        self.retry_backoff = retry_backoff
        self.out_dir = f"output/{bot_id!s}"
        os.makedirs(self.out_dir, exist_ok=True)


    async def write_results(self):
        urls = list(self.visited_urls)
        urls.sort()
        async with aiofiles.open(f"{self.out_dir}/results.json", "w", encoding="utf-8") as f:
            await f.write(json.dumps({
                "urls": urls,
                "domain": self.domain,
                "pages": self.pages,
            }, indent=4))

    def _same_domain(self, url: str) -> bool:
        netloc = urlparse(url).netloc.removeprefix("www.")
        return netloc == self.domain

    async def extract_links(self, page: Page) -> list:
        """Returns normalized, same-domain absolute URLs found on the page."""
        try:
            hrefs = await page.evaluate(
                "() => Array.from(document.querySelectorAll('a[href]')).map(a => a.href)"
            )
        except Error as e:
            print(f"[-] Error extracting links from {page.url}: {e}")
            return []

        links = set()
        for href in hrefs:
            if not href or not self._same_domain(href):
                continue
            parsed = urlparse(href)
            links.add(f"{parsed.scheme}://{parsed.netloc}{parsed.path}".rstrip("/"))
        return list(links)


    async def scrape_page(self, page: Page, url: str):
        """Saves the rendered HTML and a full-page screenshot for a URL."""

        key = uuid.uuid5(uuid.NAMESPACE_URL, url)
        path = f"{self.out_dir}/{key}.png"
        try:
            await page.screenshot(path=path, full_page=True)
            html = await page.content()
            title = await page.title()
            self.pages.append({
                "domain": self.domain,
                "meta": Doc(html).meta,
                "screenshot_key": path,
                "title": title,
                "url": url,
            })
        except Error as e:
            print(f"[-] Error scraping page: {e}")
            return


    async def crawl_page(self, context: BrowserContext, url: str) -> list:
        """Fetches and scrapes one URL, retrying on failure with backoff.

        Each attempt gets a fresh page/tab, since a failed goto or a crashed
        page can leave the old one in a broken state.
        """
        last_error = None
        for attempt in range(1, self.max_retries + 2):
            async with self.bounder:
                suffix = f" (attempt {attempt}/{self.max_retries + 1})" if attempt > 1 else ""
                print(f"[+] Crawling: {url}{suffix}")
                page = await context.new_page()
                try:
                    await page.goto(url, wait_until="networkidle", timeout=30000)
                    await self.scrape_page(page, url)
                    return await self.extract_links(page)
                except Error as e:
                    last_error = e
                finally:
                    await page.close()

            if attempt <= self.max_retries:
                delay = self.retry_backoff * attempt
                print(f"[!] {url} failed ({last_error}); retrying in {delay:.0f}s")
                await asyncio.sleep(delay)

        print(f"[-] Giving up on {url} after {self.max_retries + 1} attempts")
        return []

    async def worker(self, context: BrowserContext):
        """Pulls URLs off the queue and processes them until the queue drains.

        Once max_pages is hit, remaining queued URLs are drained (marked done,
        not visited) rather than left in place, so queue.join() can still
        complete instead of waiting forever on entries nobody will process.
        """
        while True:
            url = await self.queue.get()
            try:
                async with self.visited_lock:
                    if url in self.visited_urls or self.pages_crawled >= self.max_pages:
                        continue
                    self.visited_urls.add(url)
                    self.pages_crawled += 1

                discovered_links = await self.crawl_page(context, url)

                async with self.visited_lock:
                    for link in discovered_links:
                        if link not in self.visited_urls:
                            await self.queue.put(link)
            finally:
                self.queue.task_done()

    async def run(self):
        await self.queue.put(f"https://{self.domain}")

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
            await self.write_results()


def main():
    parser = argparse.ArgumentParser(
        description="Crawl a domain and scrape (HTML + screenshot) each page, in parallel."
    )
    parser.add_argument("bot_id", help="ID of the sitemap bot")
    parser.add_argument("domain", help="Domain to crawl, e.g. example.com")
    parser.add_argument("--max-concurrency", type=int, default=10, help="Concurrent pages (default: 10)")
    parser.add_argument("--max-pages", type=int, default=100, help="Hard stop on pages crawled (default: 100)")
    parser.add_argument("--out-dir", default="output", help="Output directory (default: output)")
    parser.add_argument("--max-retries", type=int, default=2, help="Retries per failed page, beyond the first attempt (default: 2)")
    parser.add_argument("--retry-backoff", type=float, default=2.0, help="Base seconds to wait before a retry, multiplied by attempt number (default: 2.0)")
    args = parser.parse_args()

    bot = SitemapBot(
        bot_id=args.bot_id,
        domain=args.domain,
        max_concurrency=args.max_concurrency,
        max_pages=args.max_pages,
        max_retries=args.max_retries,
        retry_backoff=args.retry_backoff,
    )

    print(f"Starting parallel crawl of {bot.domain} (max {bot.max_pages} pages, concurrency {bot.max_concurrency})")
    asyncio.run(bot.run())
    print(f"Finished. Pages visited: {len(bot.visited_urls)}.")

if __name__ == "__main__":
    main()
