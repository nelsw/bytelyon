<div align="center">
<a href="https://bytelyon.com" target="_blank">
<img src="public/bytelyon-banner.png" alt="ByteLyon Logo">
</a>

![Static Badge](https://img.shields.io/badge/Laravel-13.7-red?logo=laravel&color=red)
![Static Badge](https://img.shields.io/badge/PHP-8.5-777BB4?logo=php)
![Static Badge](https://img.shields.io/badge/Python-3.14-476E99?logo=python)

[![Site Deployment Status](https://img.shields.io/endpoint?url=https%3A%2F%2Fforge.laravel.com%2Fsite-badges%2F52d486b6-7617-4880-a09c-501c69ad25f0%3Fdate%3D1%26label%3D1&style=flat)](https://forge.laravel.com/nelsw/merciful-night-3dc/3268752)

[![Project CI](https://github.com/nelsw/bytelyon/actions/workflows/ci.yml/badge.svg)](https://github.com/nelsw/bytelyon-laravel/actions/workflows/ci.yml)
[![codecov](https://codecov.io/github/nelsw/bytelyon/graph/badge.svg?token=NHfHszCkTk)](https://codecov.io/github/nelsw/bytelyon)
</div>

ByteLyon is a Web Prowler & Hunter: a Laravel + Inertia/Vue app for running
scheduled **Search** (Google SERP), **News** (RSS-driven article capture),
and **Sitemap** (site crawl) bots, backed by a horizontally-scalable fleet
of local browser workers for the actual page fetching.

## Architecture

```mermaid
flowchart TD
    Browser["Browser"] --> Ui["Inertia + Vue 3 UI"]
    Ui --> Api["Laravel controllers"]
    Api --> Postgres[("Postgres")]

    Scheduler["Laravel scheduler"] --> BotJob["BotJob::handle()"]

    BotJob -->|"search / sitemap"| Enqueue["Sqs::enqueue()"]
    BotJob -->|"news"| Rss["Rss::news() feed discovery"]
    Rss --> Articles[("articles table")]
    Rss --> Enqueue

    Enqueue --> Queue[["SQS: bytelyon-scrape-jobs"]]
    Queue --> Worker["worker fleet (Playwright +\nCloakBrowser / SeleniumBase)"]
    Worker --> S3[("S3: bytelyon-private")]
    Worker -->|"POST /api/scrape-jobs/{type}/{id}/complete"| Controller["ScrapeJobController"]
    Controller --> Postgres
    Controller -->|"sitemap: re-enqueue unseen links"| Enqueue
```

Bots are dispatched on a schedule, each doing just enough work in PHP to
know _what_ to fetch, then handing the actual browsing off to the worker
fleet over SQS. Every worker response is the same shape
(`{url, screenshot_key, content_key}`) regardless of job type — all
parsing (SERP structure, article extraction, sitemap link discovery) lives
back in Laravel, not scattered across the Python workers. See
[`docker/worker/INSTRUCTIONS.md`](docker/worker/INSTRUCTIONS.md) for the
full worker-fleet flow, including exactly what gets POSTed back per job
type.
