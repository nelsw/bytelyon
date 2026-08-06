install:
	@composer install
	@php artisan key:generate
	@npm install
	@make fresh

fresh:
	@truncate -s 0 storage/logs/browser.log storage/logs/laravel.log
	@rm -rf bootstrap/cache/* reports/*
	@php artisan optimize:clear && php artisan optimize
	@composer dump-autoload

test: fresh
	@XDEBUG_MODE=coverage ./vendor/bin/sail test --coverage-html reports/
	@open reports/dashboard.html -a safari

lint:
	@./vendor/bin/pint --parallel
	@npm i
	@npm run lint
	@npm run format
	@npm run types:check

share:
	@./vendor/bin/sail share --subdomain=bytelyon

ide:
	@php artisan ide-helper:generate
	@php artisan ide-helper:models -RWD
	@php artisan ide-helper:meta
	@php artisan ide-helper:dump
