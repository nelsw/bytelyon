from datetime import datetime

from fastapi import FastAPI

from jobs.news import NewsBot
from models.bot import Bot

app = FastAPI()


@app.get("/")
async def index():
    return {"message": "🤖"}


@app.put("/news/{id}/query/{q}/since/{dt}")
async def put_news(id: int, q: str, dt: datetime):
    b = NewsBot(bot_id=id, query=q, since=dt, max_concurrency=5, max_pages=200)
    await b.run()
    return b.articles # todo - create task

@app.put("/search/{id}/query/{q}")
async def put_search(id: int, q: str):
    return None

@app.put("/sitemap/{id}/domain/{q}")
async def put_news(id: int, q: str):
    return None
