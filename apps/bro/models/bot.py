

from datetime import datetime

from pydantic import BaseModel

from ..models import bot_type

# from models import bot_type


class Bot(BaseModel):
    id: int
    query: str
    ran_at: datetime | None = None
    type: bot_type.BotType