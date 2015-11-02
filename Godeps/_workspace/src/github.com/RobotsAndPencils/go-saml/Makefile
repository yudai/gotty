NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

default: build

build: vet
	@echo "$(OK_COLOR)==> Go Building(NO_COLOR)"
	go build ./...

init:
	go get github.com/nu7hatch/gouuid
	go get github.com/kardianos/osext
	go get github.com/stretchr/testify/assert

vet: init
	@echo "$(OK_COLOR)==> Go Vetting$(NO_COLOR)"
	go vet ./...

test: vet
	@echo "$(OK_COLOR)==> Testing$(NO_COLOR)"
	go test ./...

.PHONY: default build init test vet
