FROM alpine:3.20.6
LABEL authors="n.ryzhkov"

COPY ./auth_app /app
CMD chmox +x /app
CMD ["/app"]