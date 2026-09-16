sail := @vendor/bin/sail

#
# Server
#
build:
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1
	$(sail) build --no-cache
	$(sail) npm install
	$(sail) npm build
up:
	$(sail) up -d
	@sleep 2
	@open http://localhost:80
run: up
	$(sail) npm run dev
stop:
	$(sail) stop worker
down: stop
	$(sail) down --remove-orphans --rmi local
destroy: stop
	$(sail) down server -v --remove-orphans --rmi all

#
# DB
#
migrate:
	$(sail) artisan migrate --graceful --env=testing
	$(sail) artisan migrate --graceful
rollback:
	$(sail) artisan migrate:rollback --env=testing
	$(sail) artisan migrate:rollback
seed:
	$(sail) artisan db:seed

#
# Project
#
clean:
	@rm -rf bootstrap/cache/* public/reports/* storage/framework/sessions/*
	@truncate -s 0 storage/logs/*.log
clear:
	$(sail) artisan optimize:clear
fresh: clear
	$(sail) artisan optimize
meta:
	$(sail) artisan ide-helper:generate
	$(sail) artisan ide-helper:models
	$(sail) artisan ide-helper:meta
scan:
	$(sail) artisan brain:scan
	open http://0.0.0.0/_laravel-brain

#
# Test
#
test: clear
	@rm -rf ./reports/*
	$(sail) pint --parallel
	$(sail) test --compact --coverage --coverage-html=./reports
	@sleep 3
	@open ./reports/index.html

#
# Bots
#
work:
	$(sail) start worker
rest:
	$(sail) stop worker
queues:
	$(sail) artisan queue:work -v
update:
	(cd docker/worker && ./update.sh)
