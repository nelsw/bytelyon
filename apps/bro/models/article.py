from datetime import datetime

from pydantic import BaseModel


class Article(BaseModel):
    body: str
    description: str
    img_alt: str
    img_url: str
    keywords: str
    published_at: datetime
    publisher: str
    source: str
    title: str
    url: str
