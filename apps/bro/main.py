from datetime import datetime

from fastapi import FastAPI

from jobs.news import NewsBot
from models.bot import Bot

app = FastAPI()


@app.get("/")
async def index():
    return {"message": "🤖"}


@app.post("/bot")
async def post_bot(b: Bot):
    return b


@app.put("/news/{id}/query/{query}/since/{since}")
async def put_news(id: int, query: str, since: datetime):
    b = NewsBot(bot_id=id, query=query, since=since, max_concurrency=5, max_pages=200)
    await b.run()
    return b.articles
