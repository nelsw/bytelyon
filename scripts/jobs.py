import asyncio
from urllib.parse import quote, urlparse
from xml.etree.ElementTree import Element, fromstring

from playwright.async_api import (
    BrowserContext,
    Error,
    Locator as AsyncLocator,
    Page as AsyncPage,
    async_playwright,

)
from pytz import utc
from seleniumbase import SB

from http import *
from models import Article, Bot, Job


async def async_accept_cookies(page: AsyncPage) -> None:
    for text in ("Accept", "Accept all", "I agree"):
        button = page.get_by_role("button", name=text)
        try:
            if await button.count() > 0 and await button.first.is_visible():
                await button.first.click()
        except Error as e:
            print("failed to accept cookies", e)

class News(Job):
    def __init__(
        self,
        bot: Bot,
    ):
        super().__init__(
            headless=bot.headless,
            max_concurrency=5,
            max_retries=3,
            retry_backoff=3.0,
        )
        self.bot = bot
        self.visited_urls: set = set()

    @staticmethod
    async def fetch_xml(session: aiohttp.ClientSession, url: str) -> Element[str] | None:
        print(f"[ ] fetch_xml {url}")
        async with session.get(url) as response:
            if response.status >= 300:
                print(f"[!] fetch_xml {url} - {response.status}")
                return None

            print(f"[+] fetch_xml {url}")
            return fromstring(text=await response.text(encoding="utf-8"))

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
                    await page.goto(a.url, wait_until="domcontentloaded", timeout=5000)
                    await publish_news(self.bot, a, page)
                    print(f"[+] Appended Article: {page.url}")
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

    async def fetch(self, session: aiohttp.ClientSession, url: str) -> None:
        xml = await self.fetch_xml(session, url)
        if xml is not None:
            for e in xml.findall(".//item"):
                at: datetime = datetime.strptime(
                    e.findtext("pubDate", default=""), "%a, %d %b %Y %H:%M:%S %Z"
                ).astimezone(utc)
                if at > self.bot.last_run_at:
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


class Search(Job):
    def __init__(
        self,
        bot: Bot,
    ):
        super().__init__(
            headless=bot.headless,
            max_concurrency=5,
            max_retries=3,
            retry_backoff=3.0,
        )
        self.bot: Bot = bot
        self.similar_queries: set[str] = set()

    async def handle_sponsored_products(
        self, context: BrowserContext, locators: list[AsyncLocator]
    ) -> None:
        print(f"[ ] Handling Sponsored Products {len(locators)}")
        idx = 0
        for loc in locators:
            domain = await loc.get_attribute("data-dtld")
            merchant_id = await loc.get_attribute("data-merchant-id")
            if domain is None or merchant_id is None:
                continue
            href = await loc.locator(
                f'a[data-merchant-id="{merchant_id}"]'
            ).get_attribute("href")
            if href is not None:
                p = await context.new_page()
                await p.goto(href, wait_until="domcontentloaded", timeout=5000)
                await publish_page(self.bot, page=p, index=idx, kind="sponsored_products")
                idx += 1

    async def handle_sponsored_results(
        self, context: BrowserContext, locators: list[AsyncLocator]
    ) -> None:
        print(f"[ ] Handling Sponsored Results {len(locators)}")
        idx = 0
        for loc in locators:
            domain = await loc.get_attribute("data-pcu")
            if domain is None:
                continue
            href = await loc.get_attribute("href")
            if href is not None:
                p = await context.new_page()
                await p.goto(href, wait_until="domcontentloaded", timeout=5000)
                await publish_page(self.bot, page=p, index=idx, kind="sponsored_results")
                idx += 1

    async def handle_organic_results(
        self, context: BrowserContext, locators: list[AsyncLocator]
    ) -> None:
        print(f"[ ] Handling Organic Results {len(locators)}")
        idx = 0
        for loc in locators:
            href = await loc.locator("xpath=ancestor::a[1]").get_attribute("href")
            if href is not None:
                p = await context.new_page()
                await p.goto(href, wait_until="domcontentloaded", timeout=5000)
                await publish_page(self.bot, page=p, index=idx, kind="organic_results")
                idx += 1

    async def handle_organic_products(
        self, context: BrowserContext, p: AsyncPage
    ) -> None:
        locators = await p.locator("product-viewer-entrypoint").all()
        idx = 0
        for loc in locators:
            print(f"[ ] Handling Organic Product {idx}")
            try:
                await loc.locator("img").first.click(timeout=3000, force=True)
            except Error as e:
                print(f"[!] organic products handler failed: {idx}, {e}")
                continue

            u = await p.locator("div[data-redirect-url]").first.get_attribute(
                "data-redirect-url"
            )
            if u is not None:
                np = await context.new_page()
                await np.goto(url=u, wait_until="domcontentloaded", timeout=5000)
                await publish_page(bot=self.bot, page=p, index=idx, kind="organic_products")
                idx += 1

    async def handle_similar_queries(self, page: AsyncPage) -> None:
        top = page.locator("div[data-notify-expansion]")
        bottom = page.locator("div#botstuff").locator("a")
        for e in await top.all():
            txt = await e.get_attribute("data-q")
            if txt and len(txt) > 4:
                self.similar_queries.add(txt)
        for a in await bottom.all():
            txt = await a.text_content()
            if txt and len(txt) > 4:
                self.similar_queries.add(txt)

    async def pre_process(self):
        with SB(uc=True) as sb:
            sb.activate_cdp_mode()
            endpoint_url = sb.get_endpoint_url()
            async with async_playwright() as p:
                browser = await p.chromium.connect_over_cdp(endpoint_url)
                context = browser.contexts[0]
                page = context.pages[0]
                try:
                    _ = page.goto(
                        "https://www.google.com/ncr", wait_until="domcontentloaded"
                    )
                    await async_accept_cookies(page)
                    search_box = page.locator(
                        'textarea[name="q"], input[name="q"]'
                    ).first
                    await search_box.press_sequentially(self.bot.query, delay=40)
                    await search_box.press("Enter")
                    await page.wait_for_load_state("domcontentloaded")

                    captcha = page.locator(
                        "iframe[src*='recaptcha'], form#captcha-form"
                    )
                    if captcha.count() == 0:
                        print("[+] Captcha not detected.")
                    else:
                        print("[!] Captcha detected!")
                        try:
                            sb.cdp.gui_click_captcha()
                            await asyncio.sleep(5)
                        except Error as e:
                            print(f"Automatic captcha solve failed: {e}")

                        waited = 0
                        while await captcha.count() and waited < 5 * 60:
                            await asyncio.sleep(3)
                            waited += 3
                        if waited >= 5 * 60:
                            print("[-] Timed out waiting for the CAPTCHA to be solved.")
                            return

                    await page.wait_for_selector("#search", timeout=20000)
                    await self.handle_similar_queries(page)
                    await publish_search(
                        bot=self.bot, similar_queries=self.similar_queries, page=page
                    )

                    await self.handle_organic_products(context, page)
                    await self.handle_organic_results(
                        context, await page.locator("h3[id]").all()
                    )
                    await self.handle_sponsored_results(
                        context, await page.locator("[data-pcu]").all()
                    )
                    await self.handle_sponsored_products(
                        context, await page.locator("[data-dtld]").all()
                    )

                except Error as e:
                    print(f"Error occurred while running search bot: {e}")
                finally:
                    await browser.close()

    async def task(self, context: BrowserContext) -> None:
        pass

class Sitemap(Job):
    """Crawls a domain-breadth-first, fetching multiple pages concurrently.

    Chrome is launched via SeleniumBase's undetected cdp_driver (to reduce
    bot-detection) and driven with playwright.async_api for page automation.
    """

    def __init__(
        self,
        bot: Bot,
    ):
        super().__init__(
            headless=bot.headless,
            max_concurrency=5,
            max_retries=3,
            retry_backoff=3.0,
        )
        self.urls: set[str] = set()
        self.bot: Bot = bot
        print("[+] Initializing Sitemap Job")

    async def extract_links(self, page: AsyncPage) -> list:
        """Returns normalized, same-domain absolute URLs found on the page."""
        print(f"[ ] Extracting links from page: {page.url}")
        try:
            links = set()
            for href in await page.evaluate("() => Array.from(document.querySelectorAll('a[href]')).map(a => a.href)"):
                if href is not None and str(urlparse(href).netloc).removeprefix("www.") == self.bot.query:
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
                    await publish_page(self.bot, page)
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
