#/bin/bash
# Use Bash instead of SH
export SHELL := /bin/bash

.DEFAULT_GOAL := controll

GOPATH := $(shell go env GOPATH)

APP_PATH := sample/app

# Run the application
run:
	@echo "Running..."
	@go run -race $(APP_PATH)/main.go
