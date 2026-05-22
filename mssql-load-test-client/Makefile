IMAGE_REGISTRY ?= ghcr.io/neaj-morshad-101
IMAGE_TAG      ?= latest
IMAGE_NAME     := $(IMAGE_REGISTRY)/mssql-load-test:$(IMAGE_TAG)
.PHONY: build
build:
	go build -o bin/mssql-load-test ./...
.PHONY: run
run:
	go run main.go
.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE_NAME) .
.PHONY: docker-push
docker-push:
	docker push $(IMAGE_NAME)
.PHONY: deploy-k8s
deploy-k8s:
	kubectl apply -f k8s/
.PHONY: cleanup-k8s
cleanup-k8s:
	kubectl delete -f k8s/ --ignore-not-found
.PHONY: tidy
tidy:
	go mod tidy
.PHONY: vet
vet:
	go vet ./...
