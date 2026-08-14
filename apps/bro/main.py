from fastapi import FastAPI, status

from jobs.job import Job
from jobs.news import NewsJob
from jobs.sitemap import SitemapJob
from models.bot import Bot, Type

# load_dotenv('../../.secrets/.env.local')


app = FastAPI()


@app.get("/")
async def index():
    return {"message": "🤖"}


@app.post(path="/bots", status_code=status.HTTP_200_OK)
async def post_bots(bot: Bot):
    job: Job
    match bot.type:
        case Type.news:
            job = NewsJob(bot)
        # case Type.search:
        case Type.sitemap:
            job = SitemapJob(bot)
        case _:
            return
    await job.run()
