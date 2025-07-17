VERSION ?= v0.0.5
TIMESTAMP := $(shell date +%s)
DOCKER_BUILD_TAG := $(VERSION)-$(TIMESTAMP)

build:
	go build .

build-docker:
	GOOS=linux GOARCH=amd64 go build .
	docker build . --tag kappy4uno/userapp:$(DOCKER_BUILD_TAG) --platform linux/amd64

push-docker: build-docker
	docker image tag kappy4uno/userapp:$(DOCKER_BUILD_TAG) kappy4uno/userapp:$(VERSION)
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

update-k8s: build-docker
	docker save kappy4uno/userapp:$(DOCKER_BUILD_TAG) -o miniuserapp.tar
	minikube image load miniuserapp.tar
	rm miniuserapp.tar
	docker image tag kappy4uno/userapp:$(DOCKER_BUILD_TAG) kappy4uno/userapp:latest
	kubectl set image deployment/userapp-deployment userapp=kappy4uno/userapp:$(DOCKER_BUILD_TAG)
