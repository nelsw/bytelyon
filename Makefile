mode := local
envf = .secrets/.env.$(mode)
dock = docker compose -f infra/compose.yml --env-file $(envf)
web = $(dock) exec web sh -c

lint:
	$(foreach app,bro mux web,make -C "apps/$(app)" lint;)

env:
	for app in bro mux web; do cp $(envf) "apps/$$app/.env"; done

# Install dependencies required before building or running apps
install: env
	$(foreach app,bro mux web,make -C "apps/$(app)" install;)

build: install
	@$(dock) build --no-cache

up:
	@$(dock) up -d
	# @$(web) "npm install && npm run dev && make fresh"

down:
	@$(dock) down --remove-orphans --rmi local

test: fresh
	@make env mode=testing
	@$(web) "make test"

fresh:
	@make -C "apps/web" fresh

migrate:
	@$(web) "php artisan migrate --graceful --ansi"

rollback:
	@$(web) "php artisan migrate:rollback --ansi"

seed:
	@$(web) "php artisan db:seed --ansi"

graph:
	@tree -a -d -I "node_modules|vendor|__pycache__|.git|.*_cache|lib|.junie|inertia-devtools|views|assets|migrations|.idea|.venv"
