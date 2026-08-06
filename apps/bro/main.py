from fastapi import FastAPI

from .models.bot import Bot

app = FastAPI()


@app.get("/")
async def index():
    return {"message": "🤖"}

@app.post("/bot")
async def post_bot(bot: Bot):
    return bot

# @app.put("/news/{id}/query/{query}")
# async def put_news(id: int, query: str):
#     return {"message": f"News {id} {query}"}
#
# @app.put("/bot/{id}/query/{query}")
# async def put_news(id: int, query: str):
#     return {"message": f"News {id} {query}"}