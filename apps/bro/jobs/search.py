import asyncio
import time

import httpx
from playwright.async_api import (
    BrowserContext,
    Error,
)
from playwright.sync_api import (
    sync_playwright,
    Locator as SyncLocator,
    Page as SyncPage
)
from seleniumbase import SB

from jobs.job import Job
from models.bot import Bot
from models.page import scrape_page
from models.search import Search, Link
from services.http import put_bot, del_bot, put_search_page
from utils.utils import accept_cookies


class SearchJob(Job):
    def __init__(
            self,
            bot: Bot,
            max_concurrency: int = 5,
            max_retries: int = 3,
            retry_backoff: float | int = 3.0,
    ):
        super().__init__(
            headless=bot.headless,
            max_concurrency=max_concurrency,
            max_retries=max_retries,
            retry_backoff=retry_backoff,
        )
        self.model = Search(bot)

    async def handle_sponsored_products(self, locators: list[SyncLocator]) -> None:
        print(f"[ ] Handling Sponsored Products {len(locators)}")
        idx = 0
        for loc in locators:
            domain = loc.get_attribute("data-dtld")
            merchant_id = loc.get_attribute("data-merchant-id")
            if domain is None or merchant_id is None:
                continue
            href = loc.locator(f'a[data-merchant-id="{merchant_id}"]').get_attribute(
                "href"
            )
            if href is not None:
                await self.queue.put(Link(url=href, idx=idx, kind="sponsored_products"))
                idx += 1

    async def handle_sponsored_results(
            self, locators: list[SyncLocator]
    ) -> None:
        print(f"[ ] Handling Sponsored Results {len(locators)}")
        idx = 0
        for loc in locators:
            domain = loc.get_attribute("data-pcu")
            if domain is None:
                continue
            href = loc.get_attribute("href")
            if href is not None:
                await self.queue.put(Link(url=href, idx=idx, kind="sponsored_results"))
                idx += 1

    async def handle_organic_results(self, locators: list[SyncLocator]) -> None:
        print(f"[ ] Handling Organic Results {len(locators)}")
        idx = 0
        for loc in locators:
            href = loc.locator("xpath=ancestor::a[1]").get_attribute("href")
            if href is not None:
                await self.queue.put(Link(url=href, idx=idx, kind="organic_results"))
                idx += 1

    async def handle_organic_products(self, p: SyncPage) -> None:
        locators = p.locator("product-viewer-entrypoint").all()
        print(f"[ ] Handling Organic Products {len(locators)}")
        idx = 0
        for loc in locators:
            print(f"[ ] Handling Organic Product {idx}")
            try:
                loc.locator("img").first.click(timeout=3000, force=True)
            except Error as e:
                print(f"[!] organic products handler failed: {idx}, {e}")
                continue

            u = p.locator("div[data-redirect-url]").first.get_attribute(
                "data-redirect-url"
            )
            if u is not None:
                await self.queue.put(Link(url=u, idx=idx, kind="organic_products"))
                idx += 1

    def handle_similar_queries(self, page: SyncPage) -> None:
        top = page.locator("div[data-notify-expansion]")
        bottom = page.locator("div#botstuff").locator("a")
        print(f"[ ] Handling Similar Queries {top.count() + bottom.count()}")
        for e in top.all():
            txt = e.get_attribute("data-q")
            if txt and len(txt) > 4:
                self.model.data["similar_queries"].append(txt)
        for a in bottom.all():
            txt = a.text_content()
            if txt and len(txt) > 4:
                self.model.data["similar_queries"].append(txt)

    async def task(self, context: BrowserContext) -> None:
        while True:
            link = await self.queue.get()
            try:
                async with self.work_lock:
                    page = await context.new_page()
                    await page.goto(url=link.url, wait_until="domcontentloaded", timeout=30000)
                    scraped_page = await scrape_page(page=page, kind=link.kind, index=link.idx)
                    if scraped_page is not None:
                        self.model.pages.append(scraped_page)
            finally:
                self.queue.task_done()

    async def pre_process(self):
        with SB(uc=True) as sb:
            sb.activate_cdp_mode()
            endpoint_url = sb.get_endpoint_url()
            with sync_playwright() as p:
                browser = p.chromium.connect_over_cdp(endpoint_url)
                context = browser.contexts[0]
                page = context.pages[0]
                try:
                    _ = page.goto(
                        "https://www.google.com/ncr", wait_until="domcontentloaded"
                    )
                    accept_cookies(page)
                    search_box = page.locator(
                        'textarea[name="q"], input[name="q"]'
                    ).first
                    search_box.press_sequentially(self.model.query, delay=40)
                    search_box.press("Enter")
                    page.wait_for_load_state("domcontentloaded")

                    captcha = page.locator(
                        "iframe[src*='recaptcha'], form#captcha-form"
                    )
                    if captcha.count() == 0:
                        print("[+] Captcha not detected.")
                    else:
                        print("[!] Captcha detected!")
                        try:
                            sb.cdp.gui_click_captcha()
                            time.sleep(5)
                        except Error as e:
                            print(f"Automatic captcha solve failed: {e}")

                        waited = 0
                        while captcha.count() and waited < 5 * 60:
                            time.sleep(3)
                            waited += 3
                        if waited >= 5 * 60:
                            print("[-] Timed out waiting for the CAPTCHA to be solved.")
                            return

                    page.wait_for_selector("#search", timeout=20000)

                    self.model.set_page(page)
                    self.handle_similar_queries(page)

                    await self.handle_organic_products(page)
                    await self.handle_organic_results(page.locator("h3[id]").all())
                    await self.handle_sponsored_results(page.locator("[data-pcu]").all())
                    await self.handle_sponsored_products(page.locator("[data-dtld]").all())

                except Error as e:
                    print(f"Error occurred while running search bot: {e}")
                finally:
                    browser.close()

    async def post_process(self):
        async with httpx.AsyncClient() as c:
            tasks = [put_search_page(c, self.model.id, page) for page in self.model.pages]
            tasks.append(put_bot(c, self.model.bot_id))
            tasks.append(del_bot(c, self.model.bot_id))
            await asyncio.gather(*tasks)
