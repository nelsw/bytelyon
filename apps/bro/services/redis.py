import gzip
import json
from dataclasses import asdict
from datetime import datetime

import redis
from playwright.async_api import Page as AsyncPage

from models.bot import Bot
from utils.utils import async_scroll_to_bottom_then_top

client = redis.Redis(host="localhost", port=6379, db=13, decode_responses=False)


async def async_content(page: AsyncPage) -> bytes:
    return gzip.compress(bytes(await page.content(), "utf-8"))


async def async_screenshot(page: AsyncPage) -> bytes:
    await async_scroll_to_bottom_then_top(page)
    return gzip.compress(await page.screenshot(full_page=True))


async def publish_bot(bot: Bot, result:str = 'ok') -> None:
    client.publish("bots", json.dumps({
        'bot_id': bot.id,
        'result': result,
    }))


async def publish_page(
    bot: Bot,
    page: AsyncPage,
    index: int | None = None,
    kind: str | None = None,
) -> None:
    client.publish(
        "pages",
        json.dumps(
            {
                "bot_id": bot.id,
                "bot_type": bot.type,
                "content": async_content(page),
                "index": index,
                "kind": kind,
                "screenshot": async_screenshot(page),
                "search_id": bot.serp_id,
                "sitemap_id": bot.sitemap_id,
                "title": page.title(),
                "url": page.url,
            }
        ),
    )


async def publish_news(
    bot_id: int, published_at: datetime, title: str, page: AsyncPage
) -> None:
    client.publish(
        "news",
        json.dumps(
            {
                "bot_id": bot_id,
                "title": title,
                "url": page.url,
                "published_at": published_at,
                "content": async_content(page),
            }
        ),
    )


async def publish_search(bot: Bot, similar_queries: set[str], page: AsyncPage) -> None:
    client.publish(
        "searches",
        json.dumps(
            {
                "id": bot.serp_id,
                "bot_id": bot.id,
                "similar_queries": similar_queries,
                "query": bot.query,
                "screenshot": async_screenshot(page),
                "content": async_content(page),
            }
        ),
    )


async def publish_sitemap(bot: Bot, urls: set[str]) -> None:
    client.publish(
        "sitemaps",
        json.dumps(
            {
                "bot_id": bot.id,
                "domain": bot.query,
                "urls": sorted(urls),
            }
        ),
    )
