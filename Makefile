include .env

# helpers =========

## help: print this help message
.PHONY: test/a
test/a:
	@echo 'hi'

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# DEV =======

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@sudo /etc/init.d/postgresql start
	go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/new name=$1: create a new db migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up db migrations
.PHONY: db/migrations/new
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

# Quality Control ===========
## audit: tidy, vendor dependencies, format, vet and test all code
.PHONY: audit
audit: vendor
	@echo '=== Formatting code...'
	go fmt ./...
	@echo '=== Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo '=== Running tests...'
	go test -race -vet=off ./...

## vendor
.PHONY: vendr
vendor:
	@echo '=== Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo '=== Vendoring dependencies...'
	go mod vendor

# BUILD ===========================
current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api


# PRODUCTION
PRODUCTION_HOST_IP = '22'
SSH_CONNECT = 'greenlight@${PRODUCTION_HOST_IP}'

## prod/connect: connect to the production server
.PHONY: prod/connect
prod/connect:
	ssh ${SSH_CONNECT}

## prod/deploy: deploy the api to production
.PHONY: prod/deploy
prod/deploy:
	rsync -rP --delete ./bin/api ./migrations ${SSH_CONNECT}:~
	ssh -t ${SSH_CONNECT} 'migrate -path ~/migrations -database $${GREENLIGHT_DB_DSN} up'


## prod/config/api.service: deploy the production systemd api.service file
.PHONY: prod/config/api.service
prod/config/api.service:
	resync -P ./remote/production/api.service ${SSH_CONNECT}:~
	ssh -t ${SSH_CONNECT} '\
		sudo mv ~/api.service /etc/systemd/system/ \
		&& sudo systemctl enable api \
		&& sudo systemctl restart api \
		'