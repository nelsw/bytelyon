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

def main(
        urls: list[str],
        path: str,
        headless: bool,
) -> None:
    Path(path).mkdir(parents=True, exist_ok=True)
    browser = launch(headless=headless, humanize=True, human_preset="careful")
    page:Page = browser.new_page()
    for url in urls:
        page.goto(url)
        page.wait_for_load_state('networkidle', timeout=10_000)
        name = uuid5(NAMESPACE_URL, page.url)
        with open(f"{path}/{name}.html", "w", encoding="utf-8") as file:
            file.write(page.content())
    browser.close()


if __name__ == '__main__':
    print("Starting content 🤖 ...")
    parser = argparse.ArgumentParser(description="Run the screenshot 🤖")
    _ = parser.add_argument('-p', '--path', help='File path')
    _ = parser.add_argument('-u', '--urls', help='Page URLs', nargs='+')
    _ = parser.add_argument('--headless', help='Headless browser', action='store_true')
    args = parser.parse_args()

    print(f"Path: {args.path}")
    print(f"Headless: {args.headless}")
    print(f"URLs: {len(args.urls)}")
    for idx, u in enumerate(args.urls):
        print(f"URL: {idx} - {u}")

    try:
        main(args.urls, args.path, args.headless)
        print("Done!")
    except Exception as err:
        print(f"Error: {err}")

