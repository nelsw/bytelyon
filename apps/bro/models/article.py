from datetime import datetime, tzinfo
from datetime import datetime as dt
from urllib.parse import urlparse
from xml.etree.ElementTree import Element

from pydantic import BaseModel
from pytz import timezone

from models.doc import Doc

RFC_1123: str = "%a, %d %b %Y %H:%M:%S %Z"
UTC: tzinfo = timezone("UTC")


def from_element(element: Element, after: datetime | None) -> Article | None:
    pub: datetime = dt.strptime(
        element.findtext("pubDate", default=""), RFC_1123
    ).astimezone(tz=UTC)
    if after is not None and pub < after:
        return None

    url: str = element.findtext("link", default="")
    src: str = "Bing News"
    if url.startswith("https://news.google"):
        src = "Google News"

    return Article(
        url=url,
        published_at=pub,
        source=src,
        title=element.findtext("body", default=""),
    )


class Article(BaseModel):
    url: str
    published_at: datetime
    source: str
    title: str
    body: str | None = None
    description: str | None = None
    img_alt: str | None = None
    img_url: str | None = None
    keywords: list[str] | None = None
    publisher: str | None = None

    def with_data(self, url: str, html: str) -> Article:
        d = Doc(html)
        self.url = url
        self.description = d.description()
        self.body = d.body()
        self.img_url = d.img_url()
        self.img_alt = d.img_alt()
        self.keywords = d.keywords()

        if self.title == "":
            self.title = d.title()

        if d.source() != "":
            self.publisher = d.source()
        else:
            self.publisher = str(urlparse(url).netloc).removeprefix("www.")
        return self
