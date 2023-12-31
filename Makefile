.DEFAULT_GOAL := help

## Go test
test:
	CGO_ENABLED=0 go test  -covermode=atomic ./...

## golangci-lint
lint:
	golangci-lint run --config .golangci.yml --timeout 300s

## Go mod
go-mod:
	go mod tidy -compat=1.19 && go mod vendor

## Build osx bin
build-osx: go-mod
	GOOS=darwin GOARCH=amd64 go build -o fleet-scheduler -mod vendor
	sha256sum fleet-scheduler > fleet-scheduler_checksum.txt

## Build linux bin
build-linux: go-mod
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fleet-scheduler -mod vendor
	sha256sum fleet-scheduler > fleet-scheduler_checksum.txt


# https://gist.github.com/prwhite/8168133
# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)


TARGET_MAX_CHAR_NUM=20
## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-$(TARGET_MAX_CHAR_NUM)s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
