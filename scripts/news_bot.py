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
import os
import uuid
from abc import ABC, abstractmethod
from asyncio import Semaphore
from dataclasses import InitVar, asdict, dataclass, field
from datetime import datetime
from enum import Enum
from typing import cast, override
from urllib.parse import quote
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
    Page,
    async_playwright,
)
from pytz import timezone
from seleniumbase import (  # pyright: ignore[reportMissingTypeStubs]
    # pyright: ignore[reportUnknownVariableType]
    cdp_driver,
)


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
class Bot:
    id: int
    query: str
    headless: bool
    after: datetime | None = None

    def object_key(self, url: str, ext: str = "png") -> str:
        return f"output/{self.id}/{uuid.uuid5(uuid.NAMESPACE_URL, url)}.{ext}"


@dataclass
class AsyncJob[T](ABC):
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
        response = await session.put(
            url=f"{os.getenv('APP_URL', default='http://localhost:80')}/api/{route}",
            json=json_data,
            headers={
                "Content-Type": "application/json",
                "x-api-key": os.getenv("API_KEY", default=""),
            },
        )
        print(f"PUT {route} -> \n{json_data}\n{response.status} {await response.text()}")


    async def upload(self, body: bytes, url: str, ext: str = "png") -> str:
        key = f"{os.getenv('APP_ENV', default='output')}/{self.bot.id}/{uuid.uuid5(uuid.NAMESPACE_URL, url)}.{ext}"
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
class News(AsyncJob[Headline]):
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
                    await page.wait_for_timeout(1600)
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

if __name__ == "__main__":
    _ = load_dotenv("../.env")

    parser = argparse.ArgumentParser(description="Run a 🤖")
    _ = parser.add_argument("-i", "--id", type=int, help="ID of the bot")
    _ = parser.add_argument("-q", "--query", type=str, help="Query for the bot")
    _ = parser.add_argument(
        "-a", "--after", type=str, help="Result date start", default=None
    )
    _ = parser.add_argument("--key", type=str, help="API Auth Key", default="my-random-32-character-x-api-key")
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
        query=cast(str, args.query),
        after=after,
        headless=cast(bool, args.headless),
    )

    asyncio.run(News(bot).run())
