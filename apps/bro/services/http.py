from models.article import Article
from dataclasses import asdict
from models.page import Page
from models.sitemap import Sitemap
from models.bot import Bot
import os
from typing import Mapping, Any

from httpx import AsyncClient, Response

web_app_url = os.getenv("APP_URL", default='')
mux_app_url = os.getenv("MUX_URL", default='')

async def del_bot(c: AsyncClient, bot_id: int) -> Response:
    return await c.delete(url=f"{web_app_url}/api/bots/{bot_id}")

async def put_bot(c: AsyncClient, bot_id: int, result: str = "ok") -> Response:
    return await c.put(url=f"{web_app_url}/api/bots/{bot_id}", data={"result": result})

async def put_article(c: AsyncClient, bot_id: int, a: Article) -> Response:
    return await c.put(url=f"{web_app_url}/api/bots/{bot_id}/articles", data=asdict(a))

async def put_search(c: AsyncClient, bot_id: int, data: Mapping[str, Any]) -> Response:
    return await c.put(url=f"{web_app_url}/api/bots/{bot_id}/searches", data=data)

async def put_sitemap(c: AsyncClient, m: Sitemap) -> Response:
    return await c.put(url=f"{web_app_url}/api/bots/{m.bot_id}/sitemaps", data={
        "domain": m.domain,
        "urls": sorted(m.urls),
    })

async def put_search_page(c: AsyncClient, search_id: int, page: Page) -> Response:
    return await c.put(url=f"{web_app_url}/api/searches/{search_id}/page", data=asdict(page))

async def put_sitemap_page(c: AsyncClient, sitemap_id: int, page: Page) -> Response:
    return await c.put(url=f"{web_app_url}/api/sitemaps/{sitemap_id}/page", data=asdict(page))


