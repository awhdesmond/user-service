FROM golang:1.20 AS build

WORKDIR /app

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .
RUN make build

FROM alpine:3.20.0

WORKDIR /app
COPY --from=build /app/build/ /app
