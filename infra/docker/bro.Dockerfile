FROM python:3.14

LABEL maintainer="Connor Van Elswyk"

ARG BRO_PORT

WORKDIR /code

COPY ./apps/bro/requirements.txt /code/requirements.txt

RUN pip install --no-cache-dir --upgrade -r /code/requirements.txt

COPY ./apps/bro /code/app

CMD ["fastapi", "run", "app/main.py", "--port", "${BRO_PORT}"]
