ID := --all
sail = ./vendor/bin/sail
pint = ./vendor/bin/pint
exec = docker compose exec laravel.test sh -c

clean:
	@truncate -s 0 storage/logs/browser.log storage/logs/laravel.log
	@rm -rf bootstrap/cache/* reports/* storage/framework/sessions/*

install:
	@composer install
	@npm install

build:
	@$(sail) build --no-cache

up: build
	@$(sail) up -d
	@$(exec) "npm run dev && php artisan horizon"

down:
	@$(sail) down --remove-orphans --rmi local

destroy:
	@$(sail) down laravel.test -v --remove-orphans --rmi all

lint:
	@$(pint) --parallel
	@npm run lint

fresh:
	@$(exec) "php artisan optimize:clear && php artisan optimize"

migrate:
	@$(exec) "php artisan migrate --graceful --ansi"

rollback:
	@$(exec) "php artisan migrate:rollback --ansi"

seed:
	@$(exec) "php artisan db:seed --ansi"

tail:
	@$(exec) "php artisan pail -vvv"

work:
	@$(exec) "php artisan horizon"

forget:
	@$(exec) "php artisan horizon:forget $(ID)"

helper:
	@$(exec) "php artisan ide-helper:generate && php artisan ide-helper:models && php artisan ide-helper:meta"

test: fresh
	@XDEBUG_MODE=coverage $(sail) test --coverage-html reports/
	@open reports/dashboard.html -a safari

share:
	@$(sail) share --subdomain=bytelyon

graph:
	@tree -a -d -I "node_modules|vendor|__pycache__|.git|.*_cache|lib|.junie|inertia-devtools|views|assets|migrations|.idea|.venv"
