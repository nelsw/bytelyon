#
# Server
#
build:
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1
	@vendor/bin/sail build --no-cache
up:
	@vendor/bin/sail up -d
run: up
	@vendor/bin/sail npm install
	@vendor/bin/sail npm run dev
down:
	@vendor/bin/sail down --remove-orphans --rmi local
destroy:
	@vendor/bin/sail down server -v --remove-orphans --rmi all

fresh:
	@vendor/bin/sail artisan optimize:clear
	@vendor/bin/sail artisan optimize
clean:
	@rm -rf bootstrap/cache/* public/reports/* storage/framework/sessions/*
	@truncate -s 0 storage/logs/browser.log storage/logs/laravel.log

#
# DB
#
migrate:
	@vendor/bin/sail artisan migrate --graceful --env=testing
	@vendor/bin/sail artisan migrate --graceful
	@vendor/bin/sail artisan db:seed
rollback:
	@vendor/bin/sail artisan migrate:rollback --env=testing
	@vendor/bin/sail artisan migrate:rollback

#
# Project
#
meta:
	@vendor/bin/sail artisan ide-helper:generate
	@vendor/bin/sail artisan ide-helper:models
	@vendor/bin/sail artisan ide-helper:meta
scan:
	@vendor/bin/sail artisan brain:scan
	open http://0.0.0.0/_laravel-brain

#
# Test
#
test: fresh
	@rm -rf public/reports/*
	@vendor/bin/sail artisan config:clear
	@vendor/bin/sail test --coverage-html public/reports/
cov:
	@open public/reports/dashboard.html
	@open public/reports/index.html

#
# ꟛƒ
#
logs:
	@scripts/tail-lambda-logs.sh
news:
	./scripts/update-lambda-image.sh bytelyon-news-scraper --dir news-scraper
page:
	./scripts/update-lambda-image.sh bytelyon-page-scraper --dir page-scraper
serp:
	./scripts/update-lambda-image.sh bytelyon-serp-scraper --dir serp-scraper
