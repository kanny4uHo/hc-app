build:
	go build .

build-docker:
	GOOS=linux GOARCH=amd64 go build .
	docker build . --tag kappy4uno/hc-app --platform linux/amd64

start-docker: build-docker
	docker run hc-app
