install:
	@composer install
	@php artisan key:generate
	@make fresh

fresh:
	@truncate -s 0 storage/logs/browser.log storage/logs/laravel.log
	@rm -rf bootstrap/cache/* reports/*
	@php artisan optimize:clear && php artisan optimize
	@composer dump-autoload

test: fresh
	@echo "test not implemented"

lint:
	@echo "lint not implemented"
