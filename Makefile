SHELL := /bin/bash
ROOT := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
TITLE := $(shell basename $(ROOT))
APP_PKG := $(shell go list -m)
export BIN_OUT := $(ROOT)/bin

build:
	tool/build.sh ${BIN_OUT}/${TITLE} ${APP_PKG}/pkg/version ${APP_PKG}
	@echo -e "\n"
