ID := --all
sail = ./vendor/bin/sail
pint = ./vendor/bin/pint
exec = docker compose exec laravel.test sh -c

graph:
	@tree -a -d -I "node_modules|vendor|__pycache__|.git|.*_cache|lib|.junie|inertia-devtools|views|assets|migrations|.idea|.venv"

clean:
	@truncate -s 0 storage/logs/browser.log storage/logs/laravel.log
	@rm -rf bootstrap/cache/* reports/* storage/framework/sessions/*

lint:
	@$(pint) --parallel
	@npm run lint

install:
	@composer install
	@npm install

build:
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1
	@$(sail) build --no-cache

up:
	@$(sail) up -d

down:
	@$(sail) down --remove-orphans --rmi local

destroy:
	@$(sail) down laravel.test -v --remove-orphans --rmi all

run: up
	@$(exec) "npm run dev && php artisan horizon"

fresh:
	@$(exec) "php artisan optimize:clear && php artisan optimize"

migrate:
	@$(exec) "php artisan migrate --graceful --ansi"
	@$(exec) "php artisan migrate --env=testing --graceful --ansi"

rollback:
	@$(exec) "php artisan migrate:rollback --ansi"
	@$(exec) "php artisan migrate:rollback --env=testing --ansi"

squash:
	@$(exec) "php artisan schema:dump"
	@$(exec) "php artisan schema:dump --database=testing --prune"

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
