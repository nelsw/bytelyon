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
import asyncio
import gzip
import json
import os
import sys
import uuid
from abc import ABC, abstractmethod
from asyncio import Semaphore
from dataclasses import dataclass, field, InitVar, asdict
from datetime import datetime
from enum import Enum
from typing import Any
from urllib.parse import quote, urlparse
from xml.etree.ElementTree import Element, fromstring

import aioboto3
import aiohttp
import httpx
from aiohttp import ClientSession
from bs4 import BeautifulSoup, NavigableString, Tag
from dotenv import load_dotenv
from playwright.async_api import (
    Error,
    BrowserContext, async_playwright,
    Page as AsyncPage,
    Locator as AsyncLocator,
)
from pytz import timezone
from seleniumbase import cdp_driver, SB


class Type(str, Enum):
    news = "news"
    search = "search"
    sitemap = "sitemap"


@dataclass
class Headline:
    url: str
    published_at: str
    title: str

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
class Bot:
    id: int
    type: Type
    query: str
    last_run_at: datetime
    headless: bool
    serp_id: int = field(default_factory=int)
    sitemap_id: int = field(default_factory=int)
    blacklist: set[str] = field(default_factory=set)

    def object_key(self, url: str, ext: str = 'png') -> str:
        if self.type == Type.search and url.startswith("https://www.google.com"):
            url = f"https://www.google.com?q={quote(url)}"
        return f"output/{self.id}/{uuid.uuid5(uuid.NAMESPACE_URL, url)}.{ext}"


@dataclass
class Doc:
    html: InitVar[str]
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
            vals = set()
            if k in self.meta:
                vals = set(self.meta[k])

            for v in set(vv):
                for val in set(v.split(",")):
                    vals.add(val.strip())

            self.meta[k] = list(vals)
        print('made doc')

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
        return self.value("twitter:image:src", "twitter:image", "og:image:secure_url", "og:image:url", "og:image",
                          "image")

    def img_alt(self) -> str:
        return self.value("twitter:image:alt", "og:image:alt")

    def source(self) -> str:
        return self.value("twitter:site", "og:site_name", "og:site")

    def description(self) -> str:
        return self.value("twitter:description", "og:description", "description", "abstract")

    def keywords(self) -> list[str]:
        kw = set()
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

        unique_text = {}
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
                if isinstance(content, NavigableString) and not isinstance(content, Tag):
                    text = str(content).strip()
                    if text and text not in unique_text:
                        unique_text[text] = index
                    index += 1

        # Sort by index to preserve order
        ordered_text = sorted(unique_text.items(), key=lambda item: item[1])
        return "\n\n".join([item[0] for item in ordered_text])


@dataclass
class Job(ABC):
    bot: Bot

    max_concurrency: int = 5
    max_retries: int = 3
    retry_backoff: float | int = 3.0

    bounder: Semaphore = field(init=False)
    queue: asyncio.Queue = field(init=False)
    lock: asyncio.Lock = field(init=False)
    s3: aioboto3.Session = field(init=False)

    def __post_init__(self):
        self.bounder = asyncio.Semaphore(self.max_concurrency)
        self.queue = asyncio.Queue()
        self.lock = asyncio.Lock()
        self.s3 = aioboto3.Session()

    @staticmethod
    async def accept_cookies(page: AsyncPage) -> None:
        for text in ("Accept", "Accept all", "I agree"):
            button = page.get_by_role("button", name=text)
            try:
                if await button.count() > 0 and await button.first.is_visible():
                    await button.first.click()
            except Error as e:
                print("failed to accept cookies", e)

    @staticmethod
    def domain(s: str) -> str:
        return str(urlparse(s).netloc).removeprefix("www.")

    @staticmethod
    async def put(session: aiohttp.ClientSession, route: str, json_data: dict[str, Any]) -> None:
        await session.put(
            url=f"{os.getenv("APP_URL", default="http://localhost:80")}/api/{route}",
            json=json_data,
            headers={
                "Content-Type": "application/json",
                "x-api-key": os.getenv("API_KEY"),
            })

    @staticmethod
    async def content(page: AsyncPage) -> bytes:
        return gzip.compress(bytes(await page.content(), "utf-8"))

    async def upload(self, body: bytes, url: str, ext: str = 'png') -> str:
        if self.bot.type == Type.search and url.startswith("https://www.google.com"):
            url = f"https://www.google.com?q={quote(url)}"
        key = f"output/{self.bot.id}/{uuid.uuid5(uuid.NAMESPACE_URL, url)}.{ext}"
        async with self.s3.client("s3") as s3:
            await s3.put_object(
                Body=body,
                Bucket=os.getenv("AWS_BUCKET"),
                Key=key,
            )
        return key

    @staticmethod
    async def screenshot(page: AsyncPage) -> bytes:
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
        d = await cdp_driver.start_async(headless=self.bot.headless)
        url = d.get_endpoint_url()
        async with async_playwright() as p:
            browser = await p.chromium.connect_over_cdp(endpoint_url=url)
            try:
                workers = [
                    asyncio.create_task(coro=self.task(context=browser.contexts[0]))
                    for _ in range(self.max_concurrency)
                ]
                await self.queue.join()
                for worker_task in workers:
                    worker_task.cancel()
                await asyncio.gather(*workers, return_exceptions=True)
            finally:
                await browser.close()
        d.stop()

    async def run(self) -> None:
        try:
            print('pre_process...')
            await self.pre_process()
            print('process...')
            await self.process()
            print('post_process...')
            await self.post_process()
            print('put ok...')
            await put(f"bots/{self.bot.id}", {"result": "ok"})
        except Exception as err:
            print(f"put error: {str(err)}")
            await put(f"bots/{self.bot.id}", {"result": str(err)})

    async def send_page(self,
                        page: AsyncPage,
                        index: int | None = None,
                        kind: str | None = None,
                        ) -> None:
        json_data = {
            "title": page.title(),
            "url": page.url,
            "domain": self.domain(page.url),
            "index": index,
            "kind": kind,
            "meta": Doc(await page.content()).meta,
            "screenshot_key": self.upload(
                body=await self.screenshot(page),
                url=page.url,
            ),
        }
        if self.bot.type == Type.sitemap:
            await put(f"sitemaps/{self.bot.sitemap_id}/page", json_data)
        else:
            await put(f"searches/{self.bot.serp_id}/page", json_data)


@dataclass
class News(Job):
    visited_urls: set = field(default_factory=set)
    articles: list = field(default_factory=list)

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
            print(f'scrape... {attempt}')
            async with self.bounder:
                page = await ctx.new_page()
                try:
                    await page.goto(headline.url, wait_until="domcontentloaded", timeout=5000)
                    content = await page.content()
                    self.articles.append(Article(
                        headline, Doc(content), page.url
                    ))
                    await page.close()
                    return
                except Error as e:
                    print(f"[!] upsert error: {str(e)}")
                    if attempt <= self.max_retries:
                        delay = self.retry_backoff * attempt
                        await asyncio.sleep(delay)
                finally:
                    await page.close()

    @staticmethod
    async def fetch_xml(session: aiohttp.ClientSession, url: str) -> Element[str] | None:
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
            return None

        for e in xml.findall(".//item"):
            at: datetime = datetime.strptime(
                e.findtext("pubDate", default=""), "%a, %d %b %Y %H:%M:%S %Z"
            ).astimezone(tz=timezone("UTC"))
            if at > self.bot.last_run_at:
                print('found article')
                await self.queue.put(
                    Headline(
                        url=e.findtext("link", default=""),
                        published_at=e.findtext("pubDate", default=""),
                        title=e.findtext("title", default=""),
                    )
                )
        return None



    async def pre_process(self) -> None:
        urls = [
            f"https://www.bing.com/news/search?format=rss&q={quote(self.bot.query)}",
            f"https://news.google.com/rss/search?q={quote(self.bot.query)}&hl=en-US&gl=US&ceid=US:en",
        ]
        async with aiohttp.ClientSession() as session:
            await asyncio.gather(*[self.fetch(session, url) for url in urls])

    async def post_process(self) -> None:
        async with aiohttp.ClientSession() as session:
            await asyncio.gather(*[self.put(session, f"bots/{self.bot.id}/articles", asdict(a)) for a in self.articles ])



@dataclass
class Search(Job):
    similar_queries: set[str] = field(default_factory=set)

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
                await self.send_page(page=p, index=idx, kind="sponsored_products")
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
                await self.send_page(page=p, index=idx, kind="sponsored_results")
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
                await self.send_page(page=p, index=idx, kind="organic_results")
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
                await self.send_page(page=p, index=idx, kind="organic_products")
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
                    await put(f"bots/{self.bot.id}/searches", {
                        "data": {
                            "similar_queries": list(self.similar_queries),
                        },
                        "content": await self.upload(
                            body=bytes(await page.content(), "utf-8"),
                            url=page.url,
                            ext="html"
                        ),
                        "screenshot_key": self.upload(
                            body=await self.screenshot(page),
                            url=page.url,
                        ),
                    })
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


@dataclass
class Sitemap(Job):
    urls: set[str] = field(default_factory=set)

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
                    await self.send_page(page)
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

    async def pre_process(self):
        await self.queue.put(f"https://{self.bot.query}")

    async def post_process(self):
        await put(f"bots/{self.bot.id}/sitemaps", {
            "urls": sorted(self.urls),
        })

async def put(route: str, json_data: dict[str, Any]) -> None:

    url = f"{os.getenv("APP_URL", default="http://localhost:80")}/api/{route}"
    headers = {
        "Content-Type": "application/json",
        "x-api-key": os.getenv("API_KEY", default=""),
    }
    return


if __name__ == "__main__":
    load_dotenv()

    data = json.loads(sys.argv[1])
    data['last_run_at'] = datetime.fromisoformat(data['last_run_at'])

    bot = Bot(**data)

    job: Job
    match bot.type:
        case Type.news:
            job = News(bot)
        case Type.search:
            job = Search(bot)
        case Type.sitemap:
            job = Sitemap(bot)

    asyncio.run(job.run())
