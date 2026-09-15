# bytelyon-worker: local-machine SQS worker (serp, news, sitemap)

One consolidated worker fleet for every bot type that needs a real browser:
Search (Google SERP), News (article capture), and Sitemap (site crawl). One
image, one queue, one location — replacing what used to be per-job-type
Lambda functions plus, most recently, a serp-only worker.

## Why local workers instead of Lambda

Google (and to a lesser extent other sites) trusts residential/home-network
egress far more than any datacenter IP, Lambda's own included. Running the
actual browser on a home network, driven by jobs pulled from a shared SQS
queue, sidesteps that entirely — and any number of machines can run this
same worker against the same queue for free, horizontal, pull-based scaling
(SQS's visibility timeout guarantees only one worker processes a given
message at a time).

## Architecture

```
SearchBotJob / NewsBotJob / SitemapBotJob  (Laravel)
  -> Sqs::enqueueScrape($type, $id, $fields)
  -> SQS queue: bytelyon-scrape-jobs  (DLQ: bytelyon-scrape-jobs-dlq, maxReceive=3)
  -> worker.py (any number of machines, long-polling)
       -> handlers/serp.py    (type=serp)     launch+navigate w/ DataImpulse proxy
       -> handlers/generic.py (type=news|sitemap)  plain page.goto(url)
       -> both return {url, screenshot_key, content_key} — nothing else.
          No parsing happens in Python at all.
       -> uploads html/screenshot to s3://bytelyon-private/worker-scrapes/<type>/...
       -> POST {LARAVEL_BASE_URL}/api/scrape-jobs/{type}/{id}/complete
            Authorization: Bearer <worker-ability Sanctum token>
  -> ScrapeJobController (app/Http/Controllers/Api/ScrapeJobController.php)
       -> serp:    App\Support\Serp      parses SERP structure -> Serp::data
       -> news:    App\Data\Html\Page    parses article content -> Article
       -> sitemap: App\Support\Html      parses links -> Page rows + re-enqueues
                    unseen links (bounded by `depth`, carried in the job)
```

**Every worker's response payload is identical** regardless of job type:
`{url, screenshot_key, content_key}`. All parsing/business logic (SERP
structure, article body extraction, sitemap link discovery) lives in the
main Laravel app, in PHP, not scattered across Python handlers — one place,
one language, easy to test and iterate on without rebuilding/redeploying a
worker image.

### Two browser providers, one image

CloakBrowser Pro's license caps concurrent sessions at **5 seats**.
`handlers/serp.py` actually needs CloakBrowser's Google-specific stealth
patches and always uses it. `handlers/generic.py` (news/sitemap) doesn't —
ordinary websites are far less aggressive about bot detection — so it
defaults to **SeleniumBase's Stealthy Playwright Mode** instead
(`../images/seleniumbase-base/seleniumbase_playwright.py`, copied into this
image at build time), which has no such seat limit. That means news/sitemap
jobs never compete with serp jobs for one of the 5 coveted cloak seats, no
matter how many run concurrently.

Both providers are baked into the one image (see the Dockerfile's
multi-stage `COPY --from=` of the SeleniumBase shim). Set the _starting_
provider per job with a `provider` field in the SQS message (`"cloakbrowser"`
or `"seleniumbase"`), or fleet-wide with the `GENERIC_BROWSER_PROVIDER` env
var.

**Automatic fallback is wired up**: if the starting attempt raises (launch
or navigation failure) or comes back with a block-shaped HTTP status
(403/429/503), `handlers/generic.py` retries once with `cloakbrowser` --
never the other direction, and never past `cloakbrowser` (nothing left to
escalate to). A plain `page.goto` timeout is still treated as best-effort
(capture whatever loaded) rather than a trigger -- plenty of ordinary slow
pages hit that without being blocked at all.

At-least-once delivery: if a worker crashes or the callback POST fails, the
message is left alone (not deleted) and SQS redelivers it automatically
after the queue's visibility timeout (300s) — up to 3 attempts before it
lands in the DLQ. No retry logic needed in the worker itself.

## ⚠️ Before testing: sync code to the actual running app

The Sail dev stack (`bytelyon-server-1`) mounts `~/PhpstormProjects/bytelyon`,
**not** this checkout (`~/tmp/bytelyon`, branch `bugfix/flaky-serp`). All of
the following need to exist there too before any callback endpoint works:

- `routes/api.php` (the three `scrape-jobs/{type}/{id}/complete` routes)
- `app/Http/Controllers/Api/ScrapeJobController.php`
- `app/Http/Requests/Api/ScrapeJobCompleteRequest.php`
- `app/Support/Serp.php`
- `app/Services/SqsService.php`
- `app/Facades/Sqs.php`
- `app/Jobs/SearchBotJob.php`, `app/Jobs/NewsBotJob.php`, `app/Jobs/SitemapBotJob.php`
- `config/services.php` (`sqs` block)
- `.env` — add `SQS_SCRAPE_JOBS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/138305277395/bytelyon-scrape-jobs`
- `tests/Feature/Jobs/{Search,News,Sitemap}BotJobTest.php` (updated to mock `Sqs` instead of asserting synchronous side effects)

Easiest path is probably pushing this branch and pulling/merging it into the
other checkout, rather than copying files by hand.

## One-time setup

**1. Issue a worker API token** (once the code above is live): log in as
whichever user should own the "system scraper" identity, go to
Settings → API Tokens, create one (any name) — it's automatically issued
the `worker` ability (`ApiTokenController::store`). Copy the plaintext
token shown once.

**2. AWS resources** (already created, nothing to do unless rebuilding from
scratch):

- SQS queue: `bytelyon-scrape-jobs` (300s visibility timeout, 20s long-poll,
  4-day retention, redrive to `bytelyon-scrape-jobs-dlq` after 3 receives)
- IAM user `bytelyon-worker`, scoped to `sqs:ReceiveMessage` /
  `sqs:DeleteMessage` / `sqs:GetQueueAttributes` on that queue and
  `s3:PutObject` on `bytelyon-private/worker-scrapes/*`. Rotate its access
  key via `aws iam create-access-key --user-name bytelyon-worker` /
  `aws iam delete-access-key ...` if needed.

## Running a worker

```bash
# From this directory:
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t bytelyon-worker:arm64 --load .

docker run -d --name bytelyon-worker --restart unless-stopped \
  -e AWS_ACCESS_KEY_ID=<bytelyon-worker access key> \
  -e AWS_SECRET_ACCESS_KEY=<bytelyon-worker secret key> \
  -e AWS_DEFAULT_REGION=us-east-1 \
  -e SCRAPE_JOBS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/138305277395/bytelyon-scrape-jobs \
  -e LARAVEL_BASE_URL=http://host.docker.internal \
  -e LARAVEL_WORKER_TOKEN=<plaintext token from step 1> \
  -e GENERIC_BROWSER_PROVIDER=seleniumbase \
  bytelyon-worker:arm64

docker logs -f bytelyon-worker
```

`GENERIC_BROWSER_PROVIDER` is optional — defaults to `seleniumbase` already
(see "Two browser providers, one image" above); only set it to
`cloakbrowser` if you want news/sitemap jobs to use CloakBrowser fleet-wide
instead.

`LARAVEL_BASE_URL=http://host.docker.internal` reaches the Sail app's port
80 on the Docker host — works out of the box on Docker Desktop for Mac.
Adjust if running the worker on a genuinely different machine (needs a
real, reachable URL for the Laravel app — a public one if the worker isn't
on the same LAN).

Run this on as many machines as you want — they all pull from the same
queue and handle whatever job type comes up (serp/news/sitemap), and SQS's
visibility timeout guarantees no two workers process the same job
concurrently.

**Always stop workers before rebuilding**
(`docker stop bytelyon-worker && docker rm bytelyon-worker`) — a running
worker holds no external lock, but you'll otherwise have two containers
racing to poll the same queue with old vs new code.

## Manual testing without the full Laravel loop

Enqueue a job directly (bypassing the Bot jobs) to test a handler + S3 side
in isolation — the callback POST will 404/fail until a real record + live
route exist, which is expected; the worker just leaves the message for
redrive in that case:

```bash
# serp
aws sqs send-message --queue-url <queue> --region us-east-1 \
  --message-body '{"type":"serp","id":1,"query":"sailing blocks"}'

# news
aws sqs send-message --queue-url <queue> --region us-east-1 \
  --message-body '{"type":"news","id":1,"url":"https://www.bbc.com/news"}'

# sitemap (root of a crawl)
aws sqs send-message --queue-url <queue> --region us-east-1 \
  --message-body '{"type":"sitemap","id":1,"url":"https://bytelyon.com","depth":5}'
```

## Known trade-offs (from moving these off a single synchronous execution)

- **Sitemap freshness**: only the root URL is unconditionally re-crawled on
  every bot run. A previously-discovered page only refreshes if it's
  re-reachable via some newly-discovered path this run — it's no longer
  proactively re-crawled just because it was seen before. The persisted
  `pages` table is now the crawl's "seen" set (replacing what used to be an
  in-memory set scoped to one job execution).
- **No aggregate "N found" notification** for sitemap crawls (there's no
  single point in time that "the crawl finished" in a distributed,
  re-enqueueing design without extra completion-tracking infra). News still
  fires `BotResultsPersisted` at _enqueue_ time ("N articles queued"),
  reworded from the old "N articles found" since the actual scraping now
  happens later, asynchronously.
- **SERP link resolution**: `App\Support\Serp::resolveUrl()` only decodes
  Google's simple `/url?q=<destination>` redirect format directly from the
  query string — no live HTTP request. Confirmed against real captured SERPs
  that Google also uses an opaque `/goto?url=<encoded-token>` format for some
  results, which can't be decoded without a live request through Google's own
  redirect service; those links currently pass through unresolved
  (`domain` ends up as `google.com` rather than the true destination). The
  Python version used to resolve these via a real HTTP request through the
  same browser/proxy session; that capability didn't move to PHP. Worth
  revisiting if destination-URL accuracy matters more than the current
  title/kind/index-only fidelity for those specific results.
