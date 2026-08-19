#!/usr/bin/env -S uv run --script
#
# /// script
# requires-python = ">=3.14"
# dependencies = [
#   "boto3",
#   "requests",
#   "beautifulsoup4",
#   "playwright",
#   "seleniumbase",
# ]
# ///
import argparse
import json
import uuid
from dataclasses import dataclass, InitVar, field
from urllib.parse import urlparse

import boto3
import requests
from botocore.exceptions import ClientError
from bs4 import BeautifulSoup, Tag, NavigableString
from playwright.sync_api import sync_playwright, Error, Page, Locator, BrowserContext
from seleniumbase import SB

s3_client = boto3.client('s3')

api_url = "http://localhost:80/api"
api_headers = {
    "Content-Type": "application/json",
    "x-api-key": "",
}
app_env = "local"
blacklist = set()
bot_id = 0
query = ""
search_id = 0

@dataclass
class Config:
    app_env: str
    api_key: str
    api_url: str

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
class Search:
    id: int
    query: str
    bot_id: int
    blacklist: set[str]
    url: str = field(init=False)
    content_key: str = field(default_factory=str)
    screenshot_key: str = field(default_factory=str)
    data: dict[str, str] = field(default_factory=dict)

@dataclass
class Node:
    index: int
    kind: str
    page: InitVar[Page]
    domain: str = field(init=False)
    title: str = field(init=False)
    url: str = field(init=False)
    meta: dict[str, list[str]] = field(init=False)
    screenshot_key: str = field(init=False)

    def __post_init__(self, page: Page):
        self.title = page.title()
        self.url = page.url
        self.domain = str(urlparse(self.url).netloc).removeprefix("www.")
        self.meta = Doc(page.content()).meta
        self.screenshot_key = file_key(page.url)

def scroll_page(page: Page):
    page.evaluate("""async () => {
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

def file_key(url: str, ext: str = 'png') -> str:
    return f"{app_env}/{bot_id}/{uuid.uuid5(uuid.NAMESPACE_URL, url)}.{ext}"


def upload_content_to_s3(body: bytes, key: str):
    upload_to_s3(body, key, "text/html")


def upload_screenshot_to_s3(body: bytes, key: str):
    upload_to_s3(body, key, "image/png")


def upload_to_s3(body: bytes, key: str, ct: str):
    try:
        print(f"ℹ️  S3 Upload: {ct} - {key}")
        response = s3_client.put_object(
            Body=body,
            Bucket="bytelyon-private",
            Key=key,
            ContentType=ct,
        )
        print(f"✅  S3 Upload: {ct} - {key} - ETag: {response['ETag']}")
    except ClientError as e:
        print(f"❌ S3 Upload: {ct} - {key} - Error: {e}")


def put_request_to_api(url: str, data: dict):
    try:
        body = json.dumps(data, indent=4).encode("utf-8")
        print(f"ℹ️  API Request: {url}", body)
        response = requests.put(url, data=body, headers=api_headers)
        response.raise_for_status()  # Raises an error for bad status codes (4xx, 5xx)
        print("✅ API Request:", response.json())
    except requests.exceptions.RequestException as e:
        print(f"❌ API Request: {e}")


def handle_similar_queries(p: Page) -> list[str]:
    print(f"ℹ️  Similar Queries")
    top = [e.get_attribute("data-q") for e in p.locator("div[data-notify-expansion]").all()]
    end = [e.text_content() for e in p.locator("div#botstuff").locator("a").all()]
    result: set[str] = set()
    for t in top + end:
        if t is not None and len(t) > 4:
            result.add(t)
    print(f"✅ Similar Queries: {len(result)}")
    return sorted(list(result))


def save_page(p: Page, kind: str, index: int) -> None:
    scroll_page(p)
    key = file_key(p.url)
    upload_screenshot_to_s3(p.screenshot(full_page=True), key)
    put_request_to_api(f"searches/{search_id}/page", {
        "title": p.title,
        "url": p.url,
        "meta": Doc(p.content()).meta,
        "screenshot_key": key,
        "kind": kind,
        "index": index,
    })


def handle_sponsored_products(
    context: BrowserContext, locators: list[Locator]
) -> None:
    print(f"[ ] Handling Sponsored Products {len(locators)}")
    for idx, loc in enumerate(locators):
        domain = loc.get_attribute("data-dtld")
        merchant_id = loc.get_attribute("data-merchant-id")
        if domain is None or merchant_id is None:
            continue
        href = loc.locator(
            f'a[data-merchant-id="{merchant_id}"]'
        ).get_attribute("href")
        if href is not None:
            p = context.new_page()
            _ = p.goto(href, wait_until="domcontentloaded", timeout=5000)
            save_page(p, kind="sponsored_products", index=idx)


def handle_sponsored_results(
    context: BrowserContext, locators: list[Locator]
) -> None:
    print(f"[ ] Handling Sponsored Results {len(locators)}")
    for idx, loc in enumerate(locators):
        domain = loc.get_attribute("data-pcu")
        href = loc.get_attribute("href")
        if domain is None or href is None:
            continue

        p = context.new_page()
        _ = p.goto(href, wait_until="domcontentloaded", timeout=5000)
        save_page(p, kind="sponsored_results", index=idx)


def handle_organic_results(
    context: BrowserContext, locators: list[Locator]
) -> None:
    print(f"[ ] Handling Organic Results {len(locators)}")
    for idx, loc in enumerate(locators):
        href = loc.locator("xpath=ancestor::a[1]").get_attribute("href")
        if href is not None:
            p = context.new_page()
            _ = p.goto(href, wait_until="domcontentloaded", timeout=5000)
            save_page(p, kind="organic_results", index=idx)


def handle_organic_products(context: BrowserContext, p: Page) -> None:
    locators = p.locator("product-viewer-entrypoint").all()
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
        if u is not None:
            np = context.new_page()
            _ = np.goto(url=u, wait_until="domcontentloaded", timeout=5000)
            save_page(p, kind="organic_products", index=idx)


def main() -> None:
    with SB(uc=True) as sb:
        sb.activate_cdp_mode()
        endpoint_url = sb.get_endpoint_url()

        with sync_playwright() as p:
            browser = p.chromium.connect_over_cdp(endpoint_url)
            context = browser.contexts[0]
            page = context.pages[0]

            page.goto("https://www.google.com")
            page.wait_for_timeout(200)
            search_box = page.locator('textarea[name="q"], input[name="q"]').first
            search_box.press_sequentially(query, delay=40)
            search_box.press("Enter")
            page.wait_for_load_state("domcontentloaded")

            captcha = page.locator("iframe[src*='recaptcha'], form#captcha-form")
            if captcha.count() > 0:
                print("[!] Captcha detected!")
                try:
                    sb.solve_captcha()
                    sb.sleep(2)
                except Error as e:
                    print(f"solve_captcha failed: {e}")
                try:
                    sb.cdp.gui_click_captcha()
                    sb.sleep(2)
                except Error as e:
                    print(f"gui_click_captcha failed: {e}")

                waited = 0
                while captcha.count() and waited < 5 * 60:
                    waited += 3
                    sb.sleep(waited)
                if waited >= 5 * 60:
                    print("[-] Timed out waiting for the CAPTCHA to be solved.")
                    return

            _ = page.wait_for_selector("#search", timeout=20000)

            url = file_key(f"https://www.google.com?q={query.replace(' ', '+')}")
            src_key = file_key(url, 'html')
            img_key = file_key(url)
            upload_content_to_s3(bytes(page.content(), 'utf-8'), src_key)
            upload_screenshot_to_s3(page.screenshot(), img_key)
            put_request_to_api(f"bots/{bot_id}/searches", {
                "data": {
                    "similar_queries": handle_similar_queries(page),
                },
                "content_key": src_key,
                "screenshot_key": img_key,
            })
            handle_organic_products(context, page)
            handle_organic_results(
                context, page.locator("h3[id]").all()
            )
            handle_sponsored_results(
                context, page.locator("[data-pcu]").all()
            )
            handle_sponsored_products(
                context, page.locator("[data-dtld]").all()
            )
            put_request_to_api(f"bots/{bot_id}", {"result": "ok"})


# todo - login to google
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run a Search 🤖")
    _ = parser.add_argument("-i", "--id", type=int, help="ID of the bot", required=True)
    _ = parser.add_argument("-x", "--search_id", type=int, help="ID of the bot", required=True)
    _ = parser.add_argument("-q", "--query", type=str, help="Query for the bot", required=True)
    _ = parser.add_argument("-b", "--blacklist", type=str, help="Domain Blacklist", required=False, default="")
    _ = parser.add_argument("-k", "--key", type=str, help="ID of the bot", required=False,
                            default="my-random-32-character-x-api-key")
    args = parser.parse_args()

    api_headers["x-api-key"] = args.key
    if len(args.key) == 36:
        api_url = "https://bytelyon.com/api"
        app_env = "production"

    bot_id = args.id
    blacklist = set(args.blacklist.split(","))
    search_id = args.search_id
    query = args.query

    main()
