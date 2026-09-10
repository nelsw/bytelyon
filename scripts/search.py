#!/usr/bin/env -S uv run --script
#
# /// script
# requires-python = ">=3.14"
# dependencies = [
#   "cloakbrowser",
# ]
# ///
import argparse
from pathlib import Path
from uuid import uuid5, NAMESPACE_URL

from cloakbrowser import launch
from playwright.sync_api import Browser, Page

GOOGLE_URL = "https://www.google.com"
SCROLL_DN_JS = """async () => {
    let lastHeight = document.body.scrollHeight;

    while (true) {
        // 1. Calculate a random scroll increment (e.g., between 200px and 600px)
        const randomStep = Math.floor(Math.random() * (600 - 200 + 1)) + 200;

        // 2. Scroll down by that random step
        window.scrollBy({
            top: randomStep,
            behavior: 'smooth'
        });

        // 3. Wait for a random amount of time (e.g., between 400ms and 1000ms) to mimic human reading pauses
        const randomDelay = Math.floor(Math.random() * (1000 - 400 + 1)) + 400;
        await new Promise(resolve => setTimeout(resolve, randomDelay));

        // 4. Check if the page height has changed or if we've reached the true bottom
        const currentHeight = document.body.scrollHeight;
        const currentScrollPosition = window.scrollY + window.innerHeight;

        // If we are at the absolute bottom and the height hasn't increased, break the loop
        if (currentScrollPosition >= currentHeight && currentHeight === lastHeight) {
            // Give it one extra verification pause in case new content is lazy-loading
            await new Promise(resolve => setTimeout(resolve, 1500));
            if (document.body.scrollHeight === currentHeight) {
                break;
            }
        }

        lastHeight = currentHeight;
    }
}"""
SCROLL_UP_JS =
"""
() => {
    const scrollTicker = setInterval(() => {
        // Get current scroll position
        const currentTop = window.scrollY || document.documentElement.scrollTop;

        if (currentTop > 0) {
            // Generate a random step size (e.g., between 20 and 60 pixels)
            const randomStep = Math.floor(Math.random() * 40) + 200;

            // Calculate new position, ensuring it doesn't overshoot 0
            const nextScroll = Math.max(0, currentTop - randomStep);

            window.scrollTo(0, nextScroll);

            // Dynamically adjust the interval delay for the next step
            // Generates a random delay between 10ms and 30ms
            clearInterval(scrollTicker);
            setTimeout(smoothScrollToTop, Math.floor(Math.random() * 20) + 100);
        } else {
            // Stop executing once the top is reached
            clearInterval(scrollTicker);
        }
    }, Math.floor(Math.random() * 20) + 100);
}
"""


def main(
        query: str,
        path: str,
        headless: bool,
) -> None:
    Path(path).mkdir(parents=True, exist_ok=True)

    browser:Browser = launch(headless=headless, humanize=True, human_preset="careful")

    page:Page = browser.new_page()
    page.goto(GOOGLE_URL)
    page.locator('textarea[name="q"]').first.click()
    page.keyboard.type(query)
    page.keyboard.press('Enter')
    page.wait_for_selector('#search')

    page.evaluate(SCROLL_DN_JS)
    page.evaluate(SCROLL_UP_JS)

    name = uuid5(NAMESPACE_URL, f"{GOOGLE_URL}?q={query.replace(' ', '+')}")

    page.screenshot(path=f"{path}/{name}.png", full_page=True)

    with open(f"{path}/{name}.html", "w", encoding="utf-8") as file:
        file.write(page.content())

    browser.close()


if __name__ == '__main__':
    print("Starting Google 🤖 ...")
    parser = argparse.ArgumentParser(description="Run a Google 🤖")
    _ = parser.add_argument('-p', '--path', help='File path')
    _ = parser.add_argument('-q', '--query', help='Search query')
    _ = parser.add_argument('--headless', action='store_true', help='Headless browser')
    args = parser.parse_args()

    print(f"Query: {args.query}")
    print(f"Path: {args.path}")
    print(f"Headless: {args.headless}")

    try:
        main(args.query, args.path, args.headless)
        print("Done!")
    except Exception as err:
        print(f"Error: {err}")

