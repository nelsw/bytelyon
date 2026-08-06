import argparse
import json
import os
import time
import uuid
from urllib.parse import urlparse

SCROLL_PAGE_JS = """async () => {
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
    }"""

from playwright.sync_api import (
    BrowserContext,
    Error,
    Locator,
    Page,
    sync_playwright,
)
from seleniumbase import SB

from models.doc import Doc


class SearchBot:
    def __init__(self, bot_id: int, query: str):
        self.bot_id = bot_id
        self.query = query
        self.sponsored_products: list[dict] = []
        self.sponsored_results: list[dict] = []
        self.organic_results: list[dict] = []
        self.organic_products: list[dict] = []
        self.similar_queries: list[str] = []
        self.output_dir = f"output/{self.bot_id}"
        os.makedirs(self.output_dir, exist_ok=True)
        print(f"[ ] SearchBot: id={self.bot_id} query='{self.query}'")

    @staticmethod
    def create_page(
        url: str, html: str, title: str, key: str, idx: int, kind: str
    ) -> dict:
        d = Doc(html)
        domain = urlparse(url).netloc.removeprefix("www.")
        if title == "":
            title = d.title()
        if title == "":
            title = domain
        return {
            "domain": domain,
            "meta": d.meta,
            "screenshot_key": key,
            "title": title,
            "url": url,
            "index": idx,
            "kind": kind,
        }

    def handle_sponsored_products(
        self, context: BrowserContext, locators: list[Locator]
    ) -> None:
        print(f"[ ] Handling Sponsored Products {len(locators)}")
        for loc in locators:
            domain = loc.get_attribute("data-dtld")
            if not domain:
                continue
            merchant_id = loc.get_attribute("data-merchant-id")
            if merchant_id is None:
                continue
            href = loc.locator(f'a[data-merchant-id="{merchant_id}"]').get_attribute(
                "href"
            )
            if not href:
                continue

            p = context.new_page()
            try:
                p.goto(href, wait_until="domcontentloaded", timeout=30000)
                key = f"{self.output_dir}/{uuid.uuid5(uuid.NAMESPACE_URL, p.url)}.png"

                html = p.content()
                p.screenshot(path=key, full_page=True)
                title = p.title()
                self.sponsored_products.append(
                    self.create_page(
                        p.url,
                        html,
                        title,
                        key,
                        len(self.sponsored_products),
                        "sponsored_product",
                    )
                )
            except Error as e:
                print(f"sponsored products handler failed: {domain} {e}")
            finally:
                p.close()

    def handle_sponsored_results(
        self, context: BrowserContext, locators: list[Locator]
    ) -> None:
        print(f"[ ] Handling Sponsored Results {len(locators)}")
        for loc in locators:
            domain = loc.get_attribute("data-pcu")
            if not domain:
                continue
            href = loc.get_attribute("href")
            if not href:
                continue

            p = context.new_page()
            try:
                p.goto(href, wait_until="domcontentloaded", timeout=30000)
                key = f"{self.output_dir}/{uuid.uuid5(uuid.NAMESPACE_URL, p.url)}.png"
                html = p.content()
                p.screenshot(path=key, full_page=True)
                title = p.title()
                self.sponsored_results.append(
                    self.create_page(
                        p.url,
                        html,
                        title,
                        key,
                        len(self.sponsored_results),
                        "sponsored_result",
                    )
                )
            except Error as e:
                print(f"sponsored results handler failed: {href} {e}")
            finally:
                p.close()

    def handle_organic_results(
        self, context: BrowserContext, locators: list[Locator]
    ) -> None:
        print(f"[ ] Handling Organic Results {len(locators)}")
        for loc in locators:
            href = loc.locator("xpath=ancestor::a[1]").get_attribute("href")
            if not href:
                continue
            p = context.new_page()
            try:
                p.goto(href, wait_until="domcontentloaded", timeout=30000)
                key = f"{self.output_dir}/{uuid.uuid5(uuid.NAMESPACE_URL, p.url)}.png"
                html = p.content()
                p.screenshot(path=key, full_page=True)
                title = p.title()
                self.organic_results.append(
                    self.create_page(
                        href,
                        html,
                        title,
                        key,
                        len(self.organic_results),
                        "organic_result",
                    )
                )
            except Error as e:
                print(f"organic results handler failed: {href} {e}")
            finally:
                p.close()

    def handle_organic_products_v2(self, context: BrowserContext, p: Page) -> None:
        locators = p.locator("product-viewer-entrypoint").all()
        print(f"[ ] Handling Organic Products {len(locators)}")
        for idx, loc in enumerate(locators):
            print(f"[ ] Handling Organic Product {idx}")
            try:
                loc.locator("img").first.click(timeout=3000, force=True)
            except Error as e:
                print(f"[!] organic products handler failed: {idx}, {e}")
                continue

            u = p.locator("div[data-redirect-url]").first.get_attribute(
                "data-redirect-url"
            )
            if u is None:
                continue

            result = urlparse(u)
            if not all([result.scheme in ["http", "https"], result.netloc]):
                continue

            np = context.new_page()
            try:
                np.goto(u, wait_until="domcontentloaded", timeout=5000)
                key = f"{self.output_dir}/{uuid.uuid5(uuid.NAMESPACE_URL, np.url)}.png"
                np.screenshot(path=key, full_page=True)
                html = np.content()
                title = np.title()
                self.organic_products.append(
                    self.create_page(
                        np.url,
                        html,
                        title,
                        key,
                        len(self.organic_products),
                        "organic_product",
                    )
                )
            except Error as e:
                print(f"organic products handler failed: {u} {e}")
            finally:
                np.close()

    def handle_similar_queries(self, page: Page) -> None:
        top = page.locator("div[data-notify-expansion]")
        bottom = page.locator("div#botstuff").locator("a")
        print(f"[ ] Handling Similar Queries {top.count() + bottom.count()}")
        for e in top.all():
            txt = e.get_attribute("data-q")
            if txt and len(txt) > 4:
                self.similar_queries.append(txt)
        for a in bottom.all():
            txt = a.text_content()
            if txt and len(txt) > 4:
                self.similar_queries.append(txt)

    def accept_cookies(self, page: Page) -> None:
        for text in ("Accept all", "I agree", "Reject all"):
            button = page.get_by_role("button", name=text)
            try:
                if button.count() > 0 and button.first.is_visible():
                    button.first.click()
            except Error as e:
                print("failed to accept cookies", e)

    def run(self) -> None:
        with SB(uc=True) as sb:
            sb.activate_cdp_mode()
            endpoint_url = sb.get_endpoint_url()
            with sync_playwright() as p:
                browser = p.chromium.connect_over_cdp(endpoint_url)
                context = browser.contexts[0]
                page = context.pages[0]
                try:
                    print(f"[ ] Searching Google for: {self.query}")
                    _ = page.goto(
                        "https://www.google.com/ncr", wait_until="domcontentloaded"
                    )
                    self.accept_cookies(page)
                    search_box = page.locator(
                        'textarea[name="q"], input[name="q"]'
                    ).first
                    search_box.press_sequentially(self.query, delay=40)
                    search_box.press("Enter")
                    page.wait_for_load_state("domcontentloaded")

                    captcha = page.locator(
                        "iframe[src*='recaptcha'], form#captcha-form"
                    )
                    if captcha.count() == 0:
                        print("[+] Captcha not detected.")
                    else:
                        print("[!] Captcha detected!")
                        sb.activate_cdp_mode("https://www.google.com")
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
                    page.evaluate(SCROLL_PAGE_JS)
                    page.screenshot(path=f"{self.output_dir}/serp.png", full_page=True)
                    with open(
                        f"{self.output_dir}/serp.html", "w", encoding="utf-8"
                    ) as f:
                        f.write(page.content())

                    self.handle_organic_results(context, page.locator("h3[id]").all())
                    self.handle_sponsored_results(
                        context, page.locator("[data-pcu]").all()
                    )
                    self.handle_sponsored_products(
                        context, page.locator("[data-dtld]").all()
                    )
                    self.handle_similar_queries(page)
                    self.handle_organic_products_v2(context, page)
                except Error as e:
                    print(f"Error occurred while running search bot: {e}")
                finally:
                    with open(
                        f"{self.output_dir}/results.json", "w", encoding="utf-8"
                    ) as f:
                        json.dump(
                            {
                                "query": self.query,
                                "screenshot_key": f"{self.output_dir}/serp.png",
                                "content_key": f"{self.output_dir}/serp.html",
                                "sponsored_products": self.sponsored_products,
                                "sponsored_results": self.sponsored_results,
                                "organic_results": self.organic_results,
                                "organic_products": self.organic_products,
                                "similar_queries": self.similar_queries,
                            },
                            f,
                            ensure_ascii=False,
                            indent=4,
                        )
                        browser.close()


def main():
    parser = argparse.ArgumentParser(description="Search Google save page data.")
    parser.add_argument("bot_id", type=int, help="Unique identifier for the bot")
    parser.add_argument("query", type=str, help="The search query")
    args = parser.parse_args()

    bot = SearchBot(args.bot_id, args.query)

    bot.run()


if __name__ == "__main__":
    main()
