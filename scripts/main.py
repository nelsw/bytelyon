#!/usr/bin/env -S uv run --script
#
# /// script
# requires-python = ">=3.14"
# dependencies = [
#   "aioboto3",
#   "aiofiles",
#   "beautifulsoup4",
#   "httpx",
#   "playwright",
#   "python-dotenv",
#   "pytz",
#   "seleniumbase",
# ]
# ///
import argparse
import asyncio
import gzip
import os
import uuid
from abc import ABC, abstractmethod
from asyncio import Semaphore
from dataclasses import InitVar, asdict, dataclass, field
from datetime import datetime
from enum import Enum
from typing import cast, override
from urllib.parse import quote, urlparse
from xml.etree.ElementTree import Element, fromstring

import aioboto3
import aiohttp
from aiohttp import ClientSession
from bs4 import BeautifulSoup, Tag
from bs4.element import NavigableString
from dotenv import load_dotenv
from playwright.async_api import (
    BrowserContext,
    Error,
    Locator,
    Page,
    async_playwright,
)
from pytz import timezone
from seleniumbase import (  # pyright: ignore[reportMissingTypeStubs]
    SB,  # pyright: ignore[reportUnknownVariableType]
    cdp_driver,
)


class Type(str, Enum):
    news = "news"
    search = "search"
    sitemap = "sitemap"


@dataclass
class Headline:
    url: str
    published_at: str
    title: str

    def published_after(self, dt: datetime | None) -> bool:
        if dt is None:
            return True
        return (
            datetime.strptime(self.published_at, "%a, %d %b %Y %H:%M:%S %Z").astimezone(
                tz=timezone("UTC")
            )
            > dt
        )


@dataclass
class Doc:
    html: InitVar[str]
    soup: BeautifulSoup = field(init=False)
    meta: dict[str, list[str]] = field(default_factory=dict)

    def __post_init__(self, html: str) -> None:
        self.meta = dict[str, list[str]]()
        self.soup = BeautifulSoup(html, "html.parser")
        for tag in self.soup.find_all("meta"):
            k = tag.get("name") or tag.get("property")
            if not isinstance(k, str):
                continue

            vv = tag.get_attribute_list("content")
            if not k or not vv:
                continue

            k = k.lower()
            vals = set[str]()
            if k in self.meta:
                vals = set(self.meta[k])

            for v in set(vv):
                for val in set(v.split(",")):
                    vals.add(val.strip())

            self.meta[k] = list(vals)
        print("made doc")

    def value(self, *keys: str) -> str:
        for k in keys:
            vals = self.meta.get(k.lower())
            if vals:
                for v in vals:
                    v = v.strip()
                    if v:
                        return v
        return ""

    def title(self) -> str:
        v = self.value("twitter:title", "og:title", "title")
        if v:
            return v
        title_tag = self.soup.find("title")
        return title_tag.get_text().strip() if title_tag else ""

    def img_url(self) -> str:
        return self.value(
            "twitter:image:src",
            "twitter:image",
            "og:image:secure_url",
            "og:image:url",
            "og:image",
            "image",
        )

    def img_alt(self) -> str:
        return self.value("twitter:image:alt", "og:image:alt")

    def source(self) -> str:
        return self.value("twitter:site", "og:site_name", "og:site")

    def description(self) -> str:
        return self.value(
            "twitter:description", "og:description", "description", "abstract"
        )

    def keywords(self) -> list[str]:
        kw = set[str]()
        for opt in ["keywords", "news_keywords", "article:tag"]:
            vals = self.meta.get(opt)
            if vals is None:
                continue
            for val in vals:
                for v in val.split(","):
                    kw.add(v)

        return sorted(kw)

    def body(self) -> str:
        # Clone soup to avoid modifying the original if needed,
        # though in Go version it seems to modify the document if it falls back to 'body' or 'html'

        sel = self.soup.find("article")
        if not sel:
            sel = self.soup.find("main")
        if not sel:
            sel = self.soup.find("body")
        if not sel:
            sel = self.soup

        unique_text: dict[str, int] = {}
        # The Go code does: sel.Find("*").Contents().Each(...)
        # and checks if node.Type == html.TextNode
        # In BeautifulSoup, we can iterate over all descendants and check if they are NavigableString

        # We need to maintain order based on appearance
        all_elements = sel.find_all(True)  # True finds all tags

        # We also need to check the selection itself if it's a tag and has direct text
        elements_to_check = [sel] + all_elements

        index = 0
        for element in elements_to_check:
            for content in element.contents:
                if isinstance(content, NavigableString) and not isinstance(
                    content, Tag
                ):
                    text = str(content).strip()
                    if text and text not in unique_text:
                        unique_text[text] = index
                    index += 1

        # Sort by index to preserve order
        ordered_text = sorted(unique_text.items(), key=lambda item: item[1])
        return "\n\n".join([item[0] for item in ordered_text])


@dataclass
class Article:
    headline: InitVar[Headline]
    doc: InitVar[Doc]
    url: str

    published_at: str = field(init=False)
    title: str = field(init=False)

    body: str = field(init=False)
    description: str = field(init=False)
    keywords: list[str] = field(init=False)
    img_url: str = field(init=False)
    img_alt: str = field(init=False)
    source: str = field(init=False)
    publisher: str = field(init=False)

    def __post_init__(self, headline: Headline, doc: Doc) -> None:
        source = "Bing"
        if headline.url.startswith("https://news.google"):
            source = "Google"
        self.published_at = headline.published_at
        self.title = headline.title
        self.source = source
        self.publisher = doc.source()
        self.img_url = doc.img_url()
        self.img_alt = doc.img_alt()
        self.body = doc.body()
        self.keywords = doc.keywords()
        self.description = doc.description()


@dataclass
class BotPage:
    title: str
    url: str
    meta: dict[str, list[str]]
    screenshot_key: str
    domain: str = field(init=False)
    index: int | None = None
    kind: str | None = None
    
    def __post_init__(self) -> None:
        self.domain = str(urlparse(self.url).netloc).removeprefix("www.")


@dataclass
class Bot:
    id: int
    type: Type
    query: str
    headless: bool
    after: datetime | None = None
    serp_id: int = field(default_factory=int)
    sitemap_id: int = field(default_factory=int)
    blacklist: set[str] = field(default_factory=set)

    def object_key(self, url: str, ext: str = "png") -> str:
        if self.type == Type.search and url.startswith("https://www.google.com"):
            url = f"https://www.google.com?q={quote(url)}"
        return f"output/{self.id}/{uuid.uuid5(uuid.NAMESPACE_URL, url)}.{ext}"



@dataclass
class Job[T](ABC):
    bot: Bot

    max_concurrency: int = 5
    max_retries: int = 3
    retry_backoff: float | int = 3.0

    bounder: Semaphore = field(init=False)
    queue: asyncio.Queue[T] = field(init=False)
    lock: asyncio.Lock = field(init=False)
    s3_session: aioboto3.Session = field(default_factory=aioboto3.Session)

    def __post_init__(self):
        self.bounder = asyncio.Semaphore(self.max_concurrency)
        self.queue = asyncio.Queue()
        self.lock = asyncio.Lock()

    @staticmethod
    async def accept_cookies(page: Page) -> None:
        for text in ("Accept", "Accept all", "I agree"):
            button = page.get_by_role("button", name=text)
            try:
                if await button.count() > 0 and await button.first.is_visible():
                    await button.first.click()
            except Error as e:
                print("failed to accept cookies", e)

    @staticmethod
    async def put(
        session: ClientSession, route: str, json_data: dict[str, object | list[object]]
    ) -> None:
        _ = await session.put(
            url=f"{os.getenv('APP_URL', default='http://localhost:80')}/api/{route}",
            json=json_data,
            headers={
                "Content-Type": "application/json",
                "x-api-key": os.getenv("API_KEY", default=""),
            },
        )

    @staticmethod
    async def content(page: Page) -> bytes:
        return gzip.compress(bytes(await page.content(), "utf-8"))

    async def upload(self, body: bytes, url: str, ext: str = "png") -> str:
        if self.bot.type == Type.search and url.startswith("https://www.google.com"):
            url = f"https://www.google.com?q={quote(url)}"
        key = f"output/{self.bot.id}/{uuid.uuid5(uuid.NAMESPACE_URL, url)}.{ext}"
        async with self.s3_session.client("s3") as s3_client:  # pyright: ignore[reportUnknownMemberType]
            _ = await s3_client.put_object(
                Body=body,
                Bucket=os.getenv("AWS_BUCKET", ""),
                Key=key,
            )
        return key

    @staticmethod
    async def screenshot(page: Page) -> bytes:
        await page.evaluate("""async () => {
            await new Promise((resolve) => {
                let totalHeight = 0;
                let distance = 100;
                let timer = setInterval(() => {
                    let scrollHeight = document.body.scrollHeight;
                    window.scrollBy(0, distance);
                    totalHeight += distance;
                    if (totalHeight >= scrollHeight || totalHeight >= 10_000) {
                        window.scrollTo(0, 0);
                        clearInterval(timer);
                        resolve();
                    }
                }, 100);
            });
        }""")
        return await page.screenshot(full_page=True)

    @abstractmethod
    async def task(self, context: BrowserContext) -> None:
        pass

    async def pre_process(self) -> None:
        pass

    async def post_process(self) -> None:
        pass

    async def process(self):
        driver = await cdp_driver.start_async(headless=self.bot.headless)  # pyright: ignore[reportUnknownMemberType]
        endpoint_url = driver.get_endpoint_url()
        async with async_playwright() as p:
            browser = await p.chromium.connect_over_cdp(endpoint_url)
            try:
                workers = [
                    asyncio.create_task(coro=self.task(context=browser.contexts[0]))
                    for _ in range(self.max_concurrency)
                ]
                await self.queue.join()
                for worker_task in workers:
                    _ = worker_task.cancel()
                _ = await asyncio.gather(*workers, return_exceptions=True)
            finally:
                await browser.close()
        driver.stop()

    async def run(self) -> None:
        result = "ok"
        try:
            print("pre_process...")
            await self.pre_process()
            print("process...")
            await self.process()
            print("post_process...")
            await self.post_process()
            print("put ok...")
        except Error as err:
            print(f"put error: {err!s}")
            result = str(err)
        finally:
            async with aiohttp.ClientSession() as session:
                await self.put(session, f"bots/{self.bot.id}", {"result": result})


@dataclass
class News(Job[Headline]):
    visited_urls: set[str] = field(default_factory=set)
    articles: list[Article] = field(default_factory=list)

    @override
    async def task(self, context: BrowserContext):
        while True:
            a: Headline = await self.queue.get()
            if a.url == "chrome-error://chromewebdata/":
                continue
            try:
                async with self.lock:
                    if a.url in self.visited_urls:
                        continue
                    self.visited_urls.add(a.url)
                await self.scrape(context, a)
            finally:
                self.queue.task_done()

    async def scrape(self, ctx: BrowserContext, headline: Headline) -> None:
        for attempt in range(1, self.max_retries + 2):
            print(f"scrape... {attempt}")
            async with self.bounder:
                page = await ctx.new_page()
                try:
                    _ = await page.goto(
                        headline.url, wait_until="domcontentloaded", timeout=5000
                    )
                    content = await page.content()
                    self.articles.append(Article(headline, Doc(content), page.url))
                    await page.close()
                    return
                except Error as e:
                    print(f"[!] upsert error: {e!s}")
                    if attempt <= self.max_retries:
                        delay = self.retry_backoff * attempt
                        await asyncio.sleep(delay)
                finally:
                    await page.close()

    @staticmethod
    async def fetch_xml(
        session: aiohttp.ClientSession, url: str
    ) -> Element[str] | None:
        print(f"[ ] fetch_xml {url}")
        async with session.get(url) as response:
            if response.status >= 300:
                print(f"[!] fetch_xml {url} - {response.status}")
                return None

            print(f"[+] fetch_xml {url}")
            return fromstring(text=await response.text(encoding="utf-8"))

    async def fetch(self, session: aiohttp.ClientSession, url: str) -> None:
        xml = await self.fetch_xml(session, url)
        if xml is None:
            return

        for e in xml.findall(".//item"):
            h = Headline(
                url=e.findtext("link", default=""),
                published_at=e.findtext("pubDate", default=""),
                title=e.findtext("title", default=""),
            )
            if h.published_after(self.bot.after):
                await self.queue.put(h)
        return

    @override
    async def pre_process(self) -> None:
        urls = [
            f"https://www.bing.com/news/search?format=rss&q={quote(self.bot.query)}",
            f"https://news.google.com/rss/search?q={quote(self.bot.query)}&hl=en-US&gl=US&ceid=US:en",
        ]
        async with aiohttp.ClientSession() as session:
            _ = await asyncio.gather(*[self.fetch(session, url) for url in urls])

    @override
    async def post_process(self) -> None:
        async with aiohttp.ClientSession() as session:
            _ = await asyncio.gather(
                *[
                    self.put(session, f"bots/{self.bot.id}/articles", asdict(a))
                    for a in self.articles
                ]
            )


@dataclass
class Search(Job[str]):
    similar_queries: set[str] = field(default_factory=set)
    pages: list[BotPage] = field(default_factory=list)
    content_key: str = field(default_factory=str)
    screenshot_key: str = field(default_factory=str)
    
    async def add_page(self, page: Page, index: int, kind:str) -> None:
        self.pages.append(BotPage(
            title=await page.title(),
            url=page.url,
            meta=Doc(html=await page.content()).meta,
            screenshot_key=await self.upload(
                body=await self.screenshot(page),
                url=page.url,
            ),
            kind=kind,
            index=index,
        ))
    
    async def handle_sponsored_products(
        self, context: BrowserContext, locators: list[Locator]
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
                _ = await p.goto(href, wait_until="domcontentloaded", timeout=5000)
                await self.add_page(page=p, index=idx, kind="sponsored_products")
                idx += 1

    async def handle_sponsored_results(
        self, context: BrowserContext, locators: list[Locator]
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
                _ = await p.goto(href, wait_until="domcontentloaded", timeout=5000)
                await self.add_page(page=p, index=idx, kind="sponsored_results")
                idx += 1

    async def handle_organic_results(
        self, context: BrowserContext, locators: list[Locator]
    ) -> None:
        print(f"[ ] Handling Organic Results {len(locators)}")
        idx = 0
        for loc in locators:
            href = await loc.locator("xpath=ancestor::a[1]").get_attribute("href")
            if href is not None:
                p = await context.new_page()
                _ = await p.goto(href, wait_until="domcontentloaded", timeout=5000)
                await self.add_page(page=p, index=idx, kind="organic_results")
                idx += 1

    async def handle_organic_products(self, context: BrowserContext, p: Page) -> None:
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
                _ = await np.goto(url=u, wait_until="domcontentloaded", timeout=5000)
                await self.add_page(page=p, index=idx, kind="organic_products")
                idx += 1

    async def handle_similar_queries(self, page: Page) -> None:
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

    @override
    async def pre_process(self):
        with SB(uc=True) as sb:
            sb.activate_cdp_mode()  # pyright: ignore[reportUnknownMemberType]
            endpoint_url = cast(str, sb.get_endpoint_url())  # pyright: ignore[reportUnknownMemberType]
            async with async_playwright() as p:
                browser = await p.chromium.connect_over_cdp(endpoint_url)
                context = browser.contexts[0]
                page = context.pages[0]
                try:
                    _ = page.goto(
                        "https://www.google.com/ncr", wait_until="domcontentloaded"
                    )
                    await self.accept_cookies(page)
                    search_box = page.locator(
                        'textarea[name="q"], input[name="q"]'
                    ).first
                    await search_box.press_sequentially(self.bot.query, delay=40)
                    await search_box.press("Enter")
                    await page.wait_for_load_state("domcontentloaded")

                    captcha = page.locator(
                        "iframe[src*='recaptcha'], form#captcha-form"
                    )
                    if await captcha.count() == 0:
                        print("[+] Captcha not detected.")
                    else:
                        print("[!] Captcha detected!")
                        try:
                            sb.cdp.gui_click_captcha()  # pyright: ignore[reportUnknownMemberType]
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

                    _ = await page.wait_for_selector("#search", timeout=20000)
                    await self.handle_similar_queries(page)
                    self.content_key = await self.upload(
                        body=bytes(await page.content(), "utf-8"),
                        url=page.url,
                        ext="html",
                    )  
                    self.screenshot_key = await self.upload(
                        body=await self.screenshot(page),
                        url=page.url,
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

    @override
    async def task(self, context: BrowserContext) -> None:
        pass

    @override
    async def post_process(self) -> None:
        async with aiohttp.ClientSession() as session:
            tasks = [
                self.put(session,f"searches/{self.bot.serp_id}/page", asdict(page))
                for page in self.pages
            ]
            tasks.append(self.put(session, f"bots/{self.bot.id}/searches",
            {
                "data": {
                    "similar_queries": list(self.similar_queries),
                },
                "content_key": self.content_key,
                "screenshot_key": self.screenshot_key,
            }))
            _ = await asyncio.gather(*tasks)
        

@dataclass
class Sitemap(Job[str]):
    urls: set[str] = field(default_factory=set)
    pages: set[BotPage] = field(default_factory=set)
    
    async def extract_links(self, page: Page) -> list[str]:
        """Returns normalized, same-domain absolute URLs found on the page."""
        print(f"[ ] Extracting links from page: {page.url}")
        try:
            links = set[str]()
            exp: str = "() => Array.from(document.querySelectorAll('a[href]')).map(a => a.href)"
            hrefs: list[str | None] = cast(list[str | None], await page.evaluate(exp))
            for href in hrefs:
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

    async def crawl_page(self, context: BrowserContext, url: str) -> list[str]:
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
                    _ = await page.goto(url, wait_until="networkidle", timeout=30000)
                    self.pages.add(BotPage(
                        title=await page.title(),
                        url=page.url,
                        meta=Doc(html=await page.content()).meta,
                        screenshot_key=await self.upload(
                            body=await self.screenshot(page),
                            url=page.url,
                        ),
                    ))
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

    @override
    async def task(self, context: BrowserContext):
        while True:
            url = await self.queue.get()
            try:
                async with self.lock:
                    if url in self.urls:
                        continue
                    self.urls.add(url)

                discovered_links = await self.crawl_page(context, url)

                async with self.lock:
                    for link in discovered_links:
                        if link not in self.urls:
                            await self.queue.put(link)
            finally:
                self.queue.task_done()

    @override
    async def pre_process(self):
        await self.queue.put(f"https://{self.bot.query}")

    @override
    async def post_process(self):
        async with aiohttp.ClientSession() as session:
            tasks = [
                self.put(session, f"sitemaps/{self.bot.sitemap_id}/page", asdict(p))
                for p in self.pages
            ]
            tasks.append(self.put(session, f"bots/{self.bot.id}/sitemaps", {
                "urls": sorted(self.urls),
            }))
            _ = await asyncio.gather(*tasks)


if __name__ == "__main__":
    _ = load_dotenv("../.env")

    parser = argparse.ArgumentParser(description="Run a 🤖")
    _ = parser.add_argument("-i", "--id", type=int, help="ID of the bot")
    _ = parser.add_argument("-t", "--type", type=Type, help="Type of the bot")
    _ = parser.add_argument("-q", "--query", type=str, help="Query for the bot")
    _ = parser.add_argument(
        "-b", "--blacklist", type=set[str], help="Blacklist for the bot"
    )
    _ = parser.add_argument(
        "-a", "--after", type=str, help="Results after this date", default=None
    )
    _ = parser.add_argument(
        "--headless", action="store_true", help="Run in headless mode"
    )
    args = parser.parse_args()

    after_arg: str | None = cast(str | None, args.after)
    if after_arg is not None:
        after = datetime.fromisoformat(after_arg)
    else:
        after = None

    bot = Bot(
        id=cast(int, args.id),
        type=cast(Type, args.type),
        query=cast(str, args.query),
        blacklist=cast(set[str], args.blacklist),
        after=after,
        headless=cast(bool, args.headless),
    )

    match bot.type:
        case Type.news:
            asyncio.run(News(bot).run())
        case Type.search:
            asyncio.run(Search(bot).run())
        case Type.sitemap:
            asyncio.run(Sitemap(bot).run())
