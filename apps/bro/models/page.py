from services.s3 import put_png
import uuid
from typing import Optional
from dataclasses import dataclass, field
from dataclasses_json import dataclass_json, Undefined, config
from playwright.async_api import Error
from playwright.async_api import Page as AsyncPage

from models.doc import Doc
from utils.utils import parse_domain

NS_URL=uuid.NAMESPACE_URL

@dataclass_json(undefined=Undefined.EXCLUDE)
@dataclass
class Page:
    domain: str
    meta: object
    screenshot_key: str
    title: str
    url: str
    index: Optional[int] = None
    kind: Optional[str] = None

async def scrape_page(
        page: AsyncPage,
        index: Optional[int] = None,
        kind: Optional[str] = None,
) -> Page | None:
    try:
        url = page.url
        domain = parse_domain(url)
        key = f"output/{domain}/{uuid.uuid5(namespace=NS_URL, name=url)}.png"
        put_png(body=await page.screenshot(full_page=True), key=key)
        return Page(
            url=url,
            domain=domain,
            title=await page.title(),
            meta=Doc(await page.content()).meta,
            screenshot_key=key,
            index=index,
            kind=kind,
        )
    except Error as e:
        print(e)
        return None