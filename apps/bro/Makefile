lint:
	mypy . --fix
	ty check .
	ruff format
	ruff check --fix --unsafe-fixes

install:
	@uv run sync
	@uv pip compile pyproject.toml -o requirements.txt

test:
	./tests/main.sh
	coverage run -m pytest tests/
	coverage report
	coverage html --title "${@-coverage}"

up:
	uv run fastapi dev --port 8085