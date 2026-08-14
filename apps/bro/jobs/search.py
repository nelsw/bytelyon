import asyncio

from playwright.async_api import (
    BrowserContext,
    Error,
    async_playwright,
)
from playwright.async_api import (
    Locator as AsyncLocator,
)
from playwright.async_api import (
    Page as AsyncPage,
)
from seleniumbase import SB

from jobs.job import Job
from models.bot import Bot
from services.redis import publish_page, publish_search
from utils.utils import async_accept_cookies


class SearchJob(Job):
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
                await publish_page(
                    bot=self.bot, page=p, index=idx, kind="sponsored_products"
                )
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
                await publish_page(
                    bot=self.bot, page=p, index=idx, kind="sponsored_results"
                )
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
                await publish_page(
                    bot=self.bot, page=p, index=idx, kind="organic_results"
                )
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
                await publish_page(
                    bot=self.bot, page=p, index=idx, kind="organic_products"
                )
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
