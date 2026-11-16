FROM debian:stable-slim

WORKDIR /app
RUN mkdir -p public

COPY http_server /bin/http_server

ENV PORT=80

CMD [ "/bin/http_server" ]
