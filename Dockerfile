FROM alpine:3.20.6
LABEL authors="n.ryzhkov"

COPY ./healthcheckProject /app
CMD chmox +x /app
CMD ["/app"]