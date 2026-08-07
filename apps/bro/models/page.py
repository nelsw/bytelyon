from dataclasses import dataclass

from playwright.async_api import Error
from playwright.async_api import Page as AsyncPage

from models.doc import Doc
from utils.utils import parse_domain


async def from_playwright(page: AsyncPage) -> Page | None:
    try:
        return Page(
            url=page.url,
            domain=parse_domain(page.url),
            title=await page.title(),
            meta=Doc(await page.content()).meta,
            img=str(await page.screenshot()),
        )
    except Error as e:
        print(e)
        return None


@dataclass
class Page:
    url: str
    domain: str
    title: str
    meta: object
    img: str
    screenshot_key: str | None = None
    index: int | None = None
    kind: str | None = None
