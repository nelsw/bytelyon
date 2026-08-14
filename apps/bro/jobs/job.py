from services.redis import publish_bot
import asyncio
from abc import ABC
from asyncio import Semaphore
from dataclasses import dataclass, field

from playwright.async_api import BrowserContext, async_playwright
from seleniumbase import cdp_driver

from models.bot import Bot


@dataclass
class Job(ABC):

    bot: Bot
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

    async def process(self) -> None:
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

    async def task(self, context: BrowserContext) -> None:
        pass

    async def pre_process(self) -> None:
        pass

    async def post_process(self) -> None:
        pass

    async def run(self):
        await self.pre_process()
        await self.process()
        await self.post_process()
        await publish_bot(bot=self.bot)
