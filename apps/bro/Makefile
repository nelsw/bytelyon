# format:
	# ruff check . scripts --fix
	# ruff format . scripts

lint:
	mypy .
	ty check .
	ruff check .
	ruff format . --check

install:
	@uv pip compile apps/bro/pyproject.toml -o apps/bro/requirements.txt

test:
	coverage run -m pytest tests/
	coverage report
	coverage html --title "${@-coverage}"
