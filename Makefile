GITCOMMIT?=$(shell git describe --dirty --always)
CGO_ENABLED?=0
BINARY:=server
LDFLAGS:="-s -w -X github.com/awhdesmond/user-service/pkg/common.GitCommit=$(GITCOMMIT)"

CONTAINER_IMAGE?=user-service_api
CONTAINER_REGISTRY?=974860574511.dkr.ecr.eu-west-1.amazonaws.com
CONTAINER_REPOSITORY?=user-service
IMAGE_TAG?=$(GITCOMMIT)

.PHONY: build clean test db

build:
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags=$(LDFLAGS) -o build/server cmd/server/*.go

test:
	go test ./... -short -timeout 120s -race -count 1 -v

test-coverage:
	go test ./... -short -timeout 120s -race -count 1 -v -coverprofile=coverage.out
	go tool cover -html=coverage.out

db:
	./scripts/clean-db.sh postgres
	./scripts/migrate-db.sh postgres

test-db:
	docker exec user-service-postgres-1 \
		psql -U postgres -c 'CREATE DATABASE postgres_test WITH OWNER postgres' || true
	./scripts/clean-db.sh postgres_test
	./scripts/migrate-db.sh postgres_test

docker:
	docker build -t $(CONTAINER_IMAGE):$(GITCOMMIT) .
	docker tag $(CONTAINER_IMAGE):$(GITCOMMIT) $(CONTAINER_REGISTRY)/$(CONTAINER_REPOSITORY):$(IMAGE_TAG)

docker-push:
	aws ecr get-login-password --region eu-west-1 --profile terraform \
		| docker login --username AWS --password-stdin $(CONTAINER_REGISTRY)
	docker push $(CONTAINER_REGISTRY)/$(CONTAINER_REPOSITORY):$(IMAGE_TAG)

clean:
	rm -rf build cover.html coverage.out
