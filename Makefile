mode := local
envf = .secrets/.env.$(mode)
dock = docker compose -f infra/compose.yml --env-file $(envf)
bro = $(dock) exec bro sh -c
mgr = $(dock) exec mgr sh -c
web = $(dock) exec web sh -c

all: down build up

lint:
	@$(MAKE) -C apps/bro lint
	# todo - mgr
	@$(MAKE) -C apps/web lint

# Install dependencies required before building or running apps
install:
	@php -r "copy('$(envf)', 'apps/web/.env');"
	@$(MAKE) -C apps/bro install
	@$(MAKE) -C apps/web install

build: install
	@$(dock) build --no-cache

up:
	@$(dock) up -d
	@$(web) "npm run dev && make fresh"

down:
	@$(dock) down --remove-orphans --rmi local

test: fresh
	@rm -rf apps/web/.env
	@php -r "copy('.secrets/.env.testing', 'apps/web/.env.testing');"
	@$(bro) "make test"
	@$(mgr) "make test"
	@$(web) "make test"

fresh:
	@$(web) "make fresh"

migrate:
	@$(web) "php artisan migrate --graceful --ansi"

rollback:
	@$(web) "php artisan migrate:rollback --ansi"

seed:
	@$(web) "php artisan db:seed --ansi"
