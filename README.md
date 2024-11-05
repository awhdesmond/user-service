# User Service

This service provides a HTTP-based API for managing users' date of birth.

A description of the code can be found in the [`docs`](./docs/architecture.md).

## Dependencies

| Dependency | Version |
| ---------- | ------- |
| Go         | 1.20    |
| Postgres   | 16.2    |
| Redis      | 7.0     |
| Flyway     | 10.0    |

## Getting Started

```bash
# Deploy Postgres, Redis, Swagger OpenAPI
# and HTTP API using docker-compose
docker compose up -d --build

# Perform SQL migrations on postgres
brew install flyway
make db

# Run simple queries against HTTP API
./scripts/simple-query.sh

# Rebuild container image when you made changes
docker compose up -d --no-deps --build api
```

To run binary on local machine:

```bash
make build

# or use direnv (https://direnv.net/)
cp .envrc.template .envrc; export $(cat .envrc | xargs)
./build/server
```

## Environment Variables

| Environment Variable         | Description                                           |
| ---------------------------- | ----------------------------------------------------- |
| USERS_SVC_HOST               | Host to expose HTTP API server                        |
| USERS_SVC_METRICS_PORT       | Port to export HTTP API server                        |
| USERS_SVC_LOG_LEVEL          | Log Level                                             |
| USERS_SVC_CORS_ORIGIN        | CORS Origin                                           |
| USERS_SVC_POSTGRES_HOST      | Postgres Host                                         |
| USERS_SVC_POSTGRES_PORT      | Postgres Port                                         |
| USERS_SVC_POSTGRES_USERNAME  | Postgres Username                                     |
| USERS_SVC_POSTGRES_PASSWORD  | Postgres Password                                     |
| USERS_SVC_POSTGRES_DATABASE  | Postgres Database                                     |
| USERS_SVC_REDIS_URI          | Redis URI                                             |
| USERS_SVC_REDIS_PASSWORD     | Redis Password                                        |
| USERS_SVC_REDIS_CLUSTER_MODE | Redis Cluster Mode. Use non-empty string to enable it |


## Testing

Run the following commands to bootstrap the test database and run unit tests.

```bash
make test-db
make test
```

## Docker

Build docker image and push them to the repository.

```bash
export CONTAINER_REGISTRY=<REGISTRY_URL>
export CONTAINER_REPOSITORY=<REPOSITORY>

make docker
make docker-push
```

## Data Schema & Flyway

Flyway is an open-source database-migration tool that helps us to version control our data schemas.

```sql
CREATE TABLE users (
    "username" TEXT NOT NULL,
    "date_of_birth" TIMESTAMP NOT NULL,
    constraint users_pk primary key (username)
);
```

> We store `date_of_birth` using UTC timezone.

## Swagger OpenAPI

View the OpenAPI spec for this service at http://localhost:3000.

## GitHub Actions (CI/CD)

View the GitHub Actions Workflows (CI Pipelines) under `.github` directory.
