SHELL := /usr/bin/env bash

# Get the root directory for make
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

all: gofmt golint govet build test

deps-update:
	@echo "Running go mod tidy & vendor"
	go mod tidy && \
	go mod vendor

gofmt:
	@echo "Running gofmt"
	go fmt $(ROOT_DIR)/pkg
	go fmt $(ROOT_DIR)/main.go

golint:
	@echo "Running golint"
	golint $(ROOT_DIR)/pkg/...
	golint $(ROOT_DIR)/main.go

govet:
	@echo "Running govet"
	go vet $(ROOT_DIR)/pkg/
	go vet $(ROOT_DIR)/main.go

build:
	@echo "Building go binary"
	go build -o mcmaker $(ROOT_DIR)/main.go

test:
	@echo "Running go tests"
	go test $(ROOT_DIR)/pkg

clean:
	@echo "Cleaning binary"
	rm mcmaker
