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


from datetime import UTC, datetime


def get_datetime_utc() -> datetime:
    return datetime.now(UTC)


async def async_scroll_to_bottom_then_top(page: AsyncPage):
    await page.evaluate(SCROLL_PAGE_JS)


def scroll_to_bottom_then_top(page: SyncPage):
    page.evaluate(SCROLL_PAGE_JS)


def accept_cookies(page: SyncPage) -> None:
    for text in ("Accept all", "I agree", "Reject all"):
        button = page.get_by_role("button", name=text)
        try:
            if button.count() > 0 and button.first.is_visible():
                button.first.click()
        except Error as e:
            print("failed to accept cookies", e)


def handle_press_and_hold(page: SyncPage) -> None:
    if "westmarine.com" not in page.url:
        return
    print(page.content())
    print("[ ] handle_press_and_hold", page.url)

    img = page.locator("img[class=px-captcha]")
    if img.count() and img.first.is_visible():
        img.first.click(delay=5000)
        return

    for text in ("PRESS AND HOLD", "press and hold"):
        locator = page.get_by_role("button", name=text)
        try:
            if locator.count() and locator.first.is_visible():
                locator.first.click(delay=5_000)
                print("[+] handle_press_and_hold", text)
        except Error as e:
            print("[!] handle_press_and_hold", text, e)


def press_escape_key(page: SyncPage) -> None:
    page.keyboard.press("Escape")
