lint:
	@.venv/bin/mypy . --fix
	@.venv/bin/ty check .
	@ruff format
	@ruff check --fix --unsafe-fixes

install:
	@uv run sync
	@uv pip compile pyproject.toml -o requirements.txt

test:
	./tests/main.sh
	coverage run -m pytest tests/
	coverage report
	coverage html --title "${@-coverage}"

up:
	uvicorn main:app --log-config log_config.yml --env-file '.env'

upgrade:
	uv lock --upgrade
	uv sync
