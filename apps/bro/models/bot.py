from typing import Optional
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum


class Type(str, Enum):
    news = "news"
    search = "search"
    sitemap = "sitemap"


@dataclass
class Bot:
    id: int
    query: str
    last_ran_at: datetime
    type: Type
    headless: bool
    serp_id: int = field(default_factory=int)
    sitemap_id: int = field(default_factory=int)
