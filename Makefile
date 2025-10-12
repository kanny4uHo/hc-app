VERSION ?= v0.0.8
TIMESTAMP := $(shell date +%s)
DOCKER_BUILD_TAG := $(VERSION)-$(TIMESTAMP)

build:
	go build -o ./cmd/$(app)/$(app)_app ./cmd/$(app)

build-docker:
	GOOS=linux GOARCH=amd64 go build -o ./cmd/$(app)/$(app)_app ./cmd/$(app)
	docker build ./cmd/$(app) --tag kappy4uno/$(app)_app:$(DOCKER_BUILD_TAG) --platform linux/amd64

push-docker: build-docker
	docker image tag kappy4uno/$(app)_app:$(DOCKER_BUILD_TAG) kappy4uno/$(app)_app:$(VERSION)
	docker image tag kappy4uno/$(app)_app:$(DOCKER_BUILD_TAG) kappy4uno/$(app)_app:latest
	docker image push kappy4uno/$(app)_app:$(VERSION)
	docker image push kappy4uno/$(app)_app:latest

start-docker: build-docker
	docker run kappy4uno/$(app)_app

update-k8s: build-docker
	docker image tag kappy4uno/$(app)_app:$(DOCKER_BUILD_TAG) kappy4uno/$(app)_app:latest
	minikube image load kappy4uno/$(app)_app:$(DOCKER_BUILD_TAG)
	minikube image load kappy4uno/$(app)_app:latest
	kubectl set image deployment/$(app)app-deployment $(app)app=kappy4uno/$(app)_app:$(DOCKER_BUILD_TAG)
