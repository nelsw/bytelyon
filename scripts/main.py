import asyncio
import json
import sys
from datetime import datetime

from dotenv import load_dotenv

from jobs import News, Search, Sitemap
from models import Bot, Job, Type


def main() -> None:
    load_dotenv()
    data = json.loads(sys.argv[1])
    data['last_run_at'] = datetime.fromisoformat(data['last_run_at'])
    bot = Bot(**data)

    job: Job
    match bot.type:
        case Type.news:
            job = News(bot)
        case Type.search:
            job = Search(bot)
        case Type.sitemap:
            job = Sitemap(bot)
    asyncio.run(job.run())

if __name__ == "__main__":
    main()
