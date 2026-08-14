from dotenv import load_dotenv
from fastapi import FastAPI, status

from jobs.news import NewsJob
from jobs.search import SearchJob
from jobs.sitemap import SitemapJob
from models.bot import Bot, Type

load_dotenv()

app = FastAPI()


@app.get("/")
async def index():
    return {"message": "🤖"}


@app.post(path="/bots", status_code=status.HTTP_200_OK)
async def post_bots(bot: Bot):
    match bot.type:
        case Type.news:
            await NewsJob(bot).run()
        case Type.search:
            await SearchJob(bot).run()
        case Type.sitemap:
            await SitemapJob(bot).run()
