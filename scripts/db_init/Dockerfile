FROM alpine:3.20.6
LABEL authors="n.ryzhkov"

COPY ./db_init /app
CMD chmox +x /app
CMD ["/app"]