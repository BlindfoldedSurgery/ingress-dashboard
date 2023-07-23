FROM python:3.11 as builder

WORKDIR /app

ENV PYTHONUNBUFFERED 1

ADD .git/ .git/
ADD docs/ docs/
ADD requirements.txt requirements.txt
ADD mkdocs.yml mkdocs.yml

RUN pip install -r requirements.txt
RUN mkdocs build

FROM nginxinc/nginx-unprivileged:1.25.1

COPY --from=builder /app/site /usr/share/nginx/html
