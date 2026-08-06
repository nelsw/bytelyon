lint:
	mypy . --fix
	ty check .
	ruff check . --fix
	ruff format . --check

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