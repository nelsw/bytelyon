from datetime import datetime

from fastapi import FastAPI

import services.news
from models import bot

app = FastAPI()


@app.get("/")
async def index():
    return {"message": "🤖"}


@app.post("/bot")
async def post_bot(b: bot.Bot):
    return b


@app.put("/news/{id}/query/{query}/since/{since}")
async def put_news(id: int, query: str, since: datetime):
    b = services.news.NewsBot(id, query, since, 5, 200)
    await b.run()
    return b.articles
