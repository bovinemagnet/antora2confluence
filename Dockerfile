FROM ruby:3.3-alpine

RUN apk add --no-cache \
      graphviz \
      font-noto \
    && gem install --no-document \
      asciidoctor \
      asciidoctor-diagram \
      asciidoctor-kroki

WORKDIR /src
ENTRYPOINT ["asciidoctor"]
