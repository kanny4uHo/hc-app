VERSION ?= v0.0.4

build:
	go build .

build-docker:
	GOOS=linux GOARCH=amd64 go build .
	docker build . --tag kappy4uno/userapp --platform linux/amd64

push-docker: build-docker
	docker image tag kappy4uno/userapp:latest kappy4uno/userapp:$(VERSION)
	docker image push kappy4uno/userapp:$(VERSION)

start-docker: build-docker
	docker run kappy4uno/userapp

build-initdb:
	go build -o ./scripts/db_init/ ./scripts/db_init

build-docker-initdb:
	GOOS=linux GOARCH=amd64 go build -o ./scripts/db_init/ ./scripts/db_init
	docker build ./scripts/db_init --tag kappy4uno/initdb --platform linux/amd64

push-docker-initdb: build-docker-initdb
	docker image tag kappy4uno/initdb:latest kappy4uno/initdb:$(VERSION)
	docker image push kappy4uno/initdb:$(VERSION)
