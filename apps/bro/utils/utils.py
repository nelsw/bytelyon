from urllib.parse import urlparse

from playwright.async_api import Page as AsyncPage
from playwright.sync_api import (
    Error,
)
from playwright.sync_api import (
    Page as SyncPage,
)

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


def parse_domain(url: str) -> str:
    if not url:
        return ""
    return str(urlparse(url).netloc).removeprefix("www.")


async def async_scroll_to_bottom_then_top(page: AsyncPage):
    await page.evaluate(SCROLL_PAGE_JS)


def scroll_to_bottom_then_top(page: SyncPage):
    page.evaluate(SCROLL_PAGE_JS)


async def async_accept_cookies(page: AsyncPage) -> None:
    for text in ("Accept", "Accept all", "I agree"):
        button = page.get_by_role("button", name=text)
        try:
            if await button.count() > 0 and await button.first.is_visible():
                await button.first.click()
        except Error as e:
            print("failed to accept cookies", e)


def accept_cookies(page: SyncPage) -> None:
    for text in ("Accept", "Accept all", "I agree"):
        button = page.get_by_role("button", name=text)
        try:
            if button.count() > 0 and button.first.is_visible():
                button.first.click()
        except Error as e:
            print("failed to accept cookies", e)


def press_escape_key(page: SyncPage) -> None:
    page.keyboard.press("Escape")
