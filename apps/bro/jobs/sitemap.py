import asyncio
import logging
from urllib.parse import urlparse

from playwright.async_api import (
    BrowserContext,
    Error,
)
from playwright.async_api import (
    Page as AsyncPage,
)

from jobs.job import Job
from models.bot import Bot
from services.redis import publish_page, publish_sitemap
from utils.utils import parse_domain

HREF_EXPRESSION = (
    "() => Array.from(document.querySelectorAll('a[href]')).map(a => a.href)"
)

logger = logging.getLogger(__name__)


class SitemapJob(Job):
    """Crawls a domain-breadth-first, fetching multiple pages concurrently.

    Chrome is launched via SeleniumBase's undetected cdp_driver (to reduce
    bot-detection) and driven with playwright.async_api for page automation.
    """

    def __init__(
        self,
        bot: Bot,
        max_concurrency: int = 5,
        max_retries: int = 3,
        retry_backoff: float = 3.0,
    ):
        super().__init__(
            bot=bot,
            max_concurrency=max_concurrency,
            max_retries=max_retries,
            retry_backoff=retry_backoff,
        )
        self.urls: set[str] = set()
        logger.info("[+] Initializing Sitemap Job")

    async def extract_links(self, page: AsyncPage) -> list:
        """Returns normalized, same-domain absolute URLs found on the page."""
        print(f"[ ] Extracting links from page: {page.url}")
        try:
            links = set()
            for href in await page.evaluate(HREF_EXPRESSION):
                if parse_domain(href) == self.bot.query:
                    parsed = urlparse(href)
                    links.add(
                        f"{parsed.scheme}://{parsed.netloc}{parsed.path}".rstrip("/")
                    )
            print(f"[+] Extracted {len(links)} links from page: {page.url}")
            return list(links)
        except Error as e:
            print(f"[-] Error extracting links from {page.url}: {e}")
            return []

    async def crawl_page(self, context: BrowserContext, url: str) -> list:
        for attempt in range(1, self.max_retries + 2):
            async with self.bounder:
                suffix = (
                    f" (attempt {attempt}/{self.max_retries + 1})"
                    if attempt > 1
                    else ""
                )
                print(f"[+] Crawling: {url}{suffix}")
                page = await context.new_page()
                try:
                    await page.goto(url, wait_until="networkidle", timeout=30000)
                    await publish_page(bot=self.bot, page=page)
                    return await self.extract_links(page)
                except Error as e:
                    if attempt <= self.max_retries:
                        delay = self.retry_backoff * attempt
                        print(f"[!] {url} failed ({e}); retrying in {delay:.0f}s")
                        await asyncio.sleep(delay)
                finally:
                    await page.close()

        print(f"[-] Giving up on {url} after {self.max_retries + 1} attempts")
        return []

    async def task(self, context: BrowserContext):
        while True:
            url = await self.queue.get()
            try:
                async with self.work_lock:
                    if url in self.urls:
                        continue
                    self.urls.add(url)

                discovered_links = await self.crawl_page(context, url)

                async with self.work_lock:
                    for link in discovered_links:
                        if link not in self.urls:
                            await self.queue.put(link)
            finally:
                self.queue.task_done()

    async def pre_process(self):
        await self.queue.put(f"https://{self.bot.query}")

    async def post_process(self):
        await publish_sitemap(bot=self.bot, urls=self.urls)
