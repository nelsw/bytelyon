import asyncio
from abc import ABC, abstractmethod
from asyncio import Semaphore
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum

from playwright.async_api import BrowserContext, async_playwright
from seleniumbase import cdp_driver
import gzip
import json
import os
from datetime import datetime
from typing import Any

import aiohttp
from playwright.async_api import Page


@dataclass
class Article:
    url: str
    published_at: datetime
    title: str


class Type(str, Enum):
    news = "news"
    search = "search"
    sitemap = "sitemap"


@dataclass
class Bot:
    id: int
    query: str
    last_run_at: datetime
    type: Type
    headless: bool
    serp_id: int = field(default_factory=int)
    sitemap_id: int = field(default_factory=int)


@dataclass
class SearchBot:
    bot: Bot
    urls: dict[str, list] = field(default_factory=dict)
    similar_queries: set[str] = field(default_factory=set)
    def __post_init__(self):
        self.urls = {
            "sponsored_products": [],
            "sponsored_results": [],
            "organic_products": [],
            "organic_results": [],
        }

@dataclass
class Job(ABC):
    headless: bool
    max_concurrency: int
    max_retries: int
    retry_backoff: float | int

    bounder: Semaphore = field(init=False)
    queue: asyncio.Queue = field(init=False)
    work_lock: asyncio.Lock = field(init=False)

    def __post_init__(self):
        self.bounder = asyncio.Semaphore(self.max_concurrency)
        self.queue = asyncio.Queue()
        self.work_lock = asyncio.Lock()

    @abstractmethod
    async def task(self, context: BrowserContext) -> None:
        pass

    async def pre_process(self) -> None:
        pass

    async def process(self):
        d = await cdp_driver.start_async(headless=self.headless)
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
        await self.pre_process()
        await self.process()

    @staticmethod
    async def put(route: str, json_data: dict[str, Any]) -> None:
        async with aiohttp.ClientSession() as session:
            await session.put(
                url=f"{os.getenv("APP_URL", default="http://localhost:80")}/api/{route}",
                json=json_data,
                headers={
                    "Content-Type": "application/json",
                    "x-api-key": os.getenv("API_KEY"),
                })

    @staticmethod
    async def async_content(page: Page) -> bytes:
        return gzip.compress(bytes(await page.content(), "utf-8"))

    @staticmethod
    async def async_screenshot(page: Page) -> bytes:
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
        return gzip.compress(await page.screenshot(full_page=True))


    async def publish_bot(self, bot: Bot, result: str = "ok") -> None:
        await self.put(f"bots/{bot.id}", {"result": result})

    async def publish_page(self,
        bot: Bot,
        page: Page,
        index: int | None = None,
        kind: str | None = None,
    ) -> None:
        await self.put(f"bots/{bot.id}", {
            "bot_id": bot.id,
            "bot_type": bot.type,
            "content": self.async_content(page),
            "index": index,
            "kind": kind,
            "screenshot": self.async_screenshot(page),
            "search_id": bot.serp_id,
            "sitemap_id": bot.sitemap_id,
            "title": page.title(),
            "url": page.url,
        })

    async def publish_news(self, bot: Bot, article: Article, page: Page) -> None:
        await self.put(f"bots/{bot.id}/articles", {
            "bot_id": bot.id,
            "title": article.title,
            "url": page.url,
            "published_at": article.published_at,
            "content": self.async_content(page),
        })

    async def publish_search(self, bot: Bot, similar_queries: set[str], page: Page) -> None:
        await self.put(f"bots/{bot.id}/searches", {
            "id": bot.serp_id,
            "bot_id": bot.id,
            "similar_queries": similar_queries,
            "query": bot.query,
            "screenshot": self.async_screenshot(page),
            "content": self.async_content(page),
        })

    async def publish_sitemap(self, bot: Bot, urls: set[str]) -> None:
        await self.put(f"bots/{bot.id}/sitemaps", {
            "bot_id": bot.id,
            "domain": bot.query,
            "urls": sorted(urls),
        })
