# /// script
# requires-python = ">=3.14"
# dependencies = [
#   "cloakbrowser",
# ]
# ///
import sys
import json

from cloakbrowser import launch
from playwright.sync_api import Page, TimeoutError

KEYWORD_ATTRS = ["property='article:tag'", "name='news_keywords'", "name='keywords'"]
DESCRIPTION_ATTRS = ["name='description'", "property='og:description'", "name='twitter:description'", "name='abstract'"]
IMG_ALT_ATTRS = ["property='og:image:alt'", "name='twitter:image:alt'"]
IMG_SRC_ATTRS = ["name='image'", "property='og:image'", "property='og:image:secure_url'", "name='twitter:image'",
                 "name='twitter:image:src'"]


def meta_content(page: Page, attrs: list[str]) -> str:
    for attr in attrs:
        try:
            text = page.locator(f"meta[{attr}]").get_attribute('content', timeout=2_500)
            if text is not None:
                return text
        except TimeoutError:
            pass
    return ''


def body(page: Page) -> str:
    for selector in ['article', 'main', 'body', 'html']:
        locator = page.locator(selector)
        if locator is not None:
            break

    text = []
    if locator is not None:
        for p in locator.locator('p').all():
            txt = p.text_content()
            if txt is not None:
                text.append(txt.strip())

    out = " ".join(text)
    out = out.replace(
        "This is a modal window. Beginning of dialog window. Escape will cancel and close the window. End of dialog window.",
        "")
    return out


def keywords(page: Page) -> list:
    words = meta_content(page, KEYWORD_ATTRS).split(',')
    for idx, word in enumerate(words):
        txt = word.strip()
        if txt is not "":
            words[idx] = txt
    words.sort()
    return words


def main():
    browser = launch(headless=True, humanize=True, human_preset="careful")
    page: Page = browser.new_page()
    data = []
    for url in sys.argv[1:]:
        try:
            page.goto(url)
            page.wait_for_load_state('domcontentloaded', timeout=5_000)
        except TimeoutError:
            pass

        data.append({
            'url': url,
            'body': body(page),
            'img_src': meta_content(page, IMG_SRC_ATTRS),
            'img_alt': meta_content(page, IMG_ALT_ATTRS),
            'description': meta_content(page, DESCRIPTION_ATTRS),
            'keywords': keywords(page),
        })

    print(json.dumps(data, indent=2))
    browser.close()


if __name__ == '__main__':
    main()
