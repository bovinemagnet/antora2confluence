FROM ruby:3.3-alpine

RUN apk add --no-cache \
      graphviz \
      font-noto \
    && gem install --no-document \
      asciidoctor \
      asciidoctor-diagram \
      asciidoctor-kroki

WORKDIR /src

# KROKI_SERVER_URL can be set via environment to point to a local Kroki instance
# e.g. docker-compose sets KROKI_SERVER_URL=http://kroki:8000
ENTRYPOINT ["asciidoctor"]
