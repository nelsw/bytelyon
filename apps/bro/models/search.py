import uuid
from dataclasses import InitVar, dataclass, field

from playwright.sync_api import Page as SyncPage

from models.bot import Bot
from models.page import Page
from services.s3 import put_html, put_png
from utils.utils import scroll_to_bottom_then_top

NS_URL = uuid.NAMESPACE_URL


@dataclass
class Link:
    url: str
    idx: int
    kind: str


@dataclass
class Search:
    bot: InitVar[Bot]

    id: int = field(init=False)
    bot_id: int = field(init=False)
    query: str = field(init=False)

    data: dict = field(default_factory=dict)
    content_key: str = field(default_factory=str)
    screenshot_key: str = field(default_factory=str)
    pages: list[Page] = field(default_factory=list)

    def __post_init__(self, bot: Bot):
        self.bot_id = bot.id
        self.id = bot.serp_id
        self.query = bot.query

    def set_page(self, page: SyncPage):

        name: str = f"output/google.com/{self.query.replace(' ', '+')}"

        self.screenshot_key = f"{name}.png"
        self.content_key = f"{name}.html"

        scroll_to_bottom_then_top(page)

        put_png(body=page.screenshot(full_page=True), key=self.screenshot_key)
        put_html(body=page.content(), key=self.content_key)
