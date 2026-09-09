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

        page.wait_for_load_state('domcontentloaded', timeout=5_000)

        body = page.locator('article').evaluate_all("els => els.map(el => Array.from(el.querySelectorAll('p')).map(p => p.innerText.trim()).join(' '))")
        if body == '':
            page.locator('body').evaluate_all("els => els.map(el => Array.from(el.querySelectorAll('p')).map(p => p.innerText.trim()).join(' '))")

        data:dict[str, Any] = {
            'url': url,
            'title': page.title(),
            'hrefs': page.locator('a').evaluate_all('els => els.map(el => el.href)'),
            'body': body,
            'meta': page.evaluate(
                """() => {
                    const ok = (e, n) => e.getAttribute(n) !== null && e.getAttribute(n) !== '';
                    return Array.from(document.querySelectorAll('meta'))
                        .filter(e => ok(e, 'content') && (ok(e, 'name') || ok(e, 'property')))
                        .map(e => ({
                            key: ok(e, 'name') ? e.getAttribute('name') : e.getAttribute('property'),
                            val: e.getAttribute('content'),
                        })).map(e => ([e.key,e.val]));
                }"""
            )
        }
        name = uuid5(NAMESPACE_URL, url)
        with open(f"{path}/{name}.json", "w", encoding="utf-8") as file:
            json.dump(data, file, indent=4)
    browser.close()


if __name__ == '__main__':
    print("Starting content 🤖 ...")
    parser = argparse.ArgumentParser(description="Run the content 🤖")
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

