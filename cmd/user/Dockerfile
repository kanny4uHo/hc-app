FROM alpine:3.20.6
LABEL authors="n.ryzhkov"

COPY ./user_app /app
CMD chmox +x /app
CMD ["/app"]