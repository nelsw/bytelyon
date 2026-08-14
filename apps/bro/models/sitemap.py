from dataclasses import InitVar, dataclass, field

from models.bot import Bot
from models.page import Page


@dataclass
class Sitemap:
    # Argument only used during construction, not saved as an attribute
    bot: InitVar[Bot]

    # Safe mutable default set & list
    urls: set[str] = field(default_factory=set)
    pages: list[Page] = field(default_factory=list)

    # Excluded from __init__, calculated post-initialization
    id: int = field(init=False)
    bot_id: int = field(init=False)
    domain: str = field(init=False)

    def __post_init__(self, bot: Bot) -> None:
        self.id = bot.sitemap_id
        self.bot_id = bot.id
        self.domain = bot.query
