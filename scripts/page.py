#!/usr/bin/env -S uv run --script
#
# /// script
# requires-python = ">=3.14"
# dependencies = [
#   "cloakbrowser",
# ]
# ///
import argparse
import json
from pathlib import Path
from typing import Any
from uuid import uuid5, NAMESPACE_URL

from cloakbrowser import launch
from playwright.sync_api import Browser, Page


BODY_JS = """els => els.map(el => Array.from(el.querySelectorAll('p')).map(p => p.innerText.trim()).join(' '))"""
META_JS = """
() => {
    const ok = (e, n) => e.getAttribute(n) !== null && e.getAttribute(n) !== '';
    return Array.from(document.querySelectorAll('meta'))
        .filter(e => ok(e, 'content') && (ok(e, 'name') || ok(e, 'property')))
        .map(e => ({
            key: ok(e, 'name') ? e.getAttribute('name') : e.getAttribute('property'),
            val: e.getAttribute('content'),
        })).map(e => ([e.key,e.val]));
}
"""


def run(urls: list[str], path: str, headless: bool) -> None:
    Path(path).mkdir(parents=True, exist_ok=True)
    browser = launch(headless=headless, humanize=True, human_preset="careful")
    page: Page = browser.new_page()
    for url in urls:
        page.goto(url)
        page.wait_for_load_state('domcontentloaded', timeout=5_000)
        name = uuid5(NAMESPACE_URL, url)
        title = page.title()

        print(f"{name} -*- {title} -*- {url}")

        meta = page.evaluate(META_JS)
        body = page.locator('article').evaluate_all(BODY_JS)
        if body == '':
            page.locator('body').evaluate_all(BODY_JS)

        data = {
            'url': url,
            'title': title,
            'body': body,
            'meta': meta,
        }

        with open(f"{path}/{name}.json", "w", encoding="utf-8") as file:
            json.dump(data, file, indent=4)

    browser.close()


if __name__ == '__main__':
    print("Starting News 🤖 ...")
    parser = argparse.ArgumentParser(description="Run the News 🤖")
    _ = parser.add_argument('-p', '--path', help='File path')
    _ = parser.add_argument('-u', '--urls', help='Page URLs', nargs='+')
    _ = parser.add_argument('--headless', help='Headless browser', action='store_true')
    args = parser.parse_args()

    print(f"Path: {args.path}")
    print(f"Headless: {args.headless}")
    print(f"URLs: {len(args.urls)}")
    for u in args.urls:
        print(f" - {u}")

    try:
        run(args.urls, args.path, args.headless)
        print("Done!")
    except Exception as err:
        print(f"Error: {err}")
