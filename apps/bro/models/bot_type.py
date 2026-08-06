from enum import Enum


class BotType(str, Enum):
    news = "news"
    search = "search"
    sitemap = "sitemap"
