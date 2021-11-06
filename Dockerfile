# First stage Build the API
FROM golang:1.17.3-alpine3.14 AS build

# update the repository and install git
RUN apk update && apk upgrade && \
    apk add --no-cache git

WORKDIR /tmp/greenlight

COPY . .

RUN GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -X main.buildTime=$(cat ./date_r) -X main.version=$(cat ./git_description)" \
    -o=./bin/linux_amd64/api ./cmd/api

# ===================
# Second stage: running the API
FROM alpine:latest

RUN apk add ca-certificates

COPY --from=build /tmp/app/bin/linux_amd64/api /app/api
COPY --from=build /tmp/app/setup.sh /app/setup.sh

WORKDIR "/app"

EXPOSE 5000

RUN /bin/sh -c ./setup.sh

# Download migrate tool
RUN curl -L --no-progress-meter https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz | tar xvz

# Run the API in prod mode
CMD ["./api", "-port=5000", "-db-dsn=postgres://postgres:hello@db/greenlight?sslmode=disable", "-env=production"]