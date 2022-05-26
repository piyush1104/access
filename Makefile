# Makefile for brytecam API
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Tools
PROTO_TOOL ?= protoc
RM ?= rm -f
GO ?= go
COPY ?= cp
GIT ?= git
DATE ?= date
DOCKER ?= docker
DOCKER_COMPOSE ?= docker-compose
BASH ?= bash
CLANG ?= clang-format
PROTOC ?= protoc

ARCH ?= amd64
ifeq ($(OS),)
	 UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
        OS := linux
	else
		ifeq ($(UNAME_S),Darwin)
			OS := darwin
		else
			OS := windows
		endif
    endif

endif

AWS_REGION := ap-south-1

# GRPC protobuf
BIN_DIR_NAME := bin
PKG_DIR_NAME := pkg
CMD_DIR_NAME := cmd
PROTO_DIR_NAME := proto
SCRIPT_DIR_NAME := scripts
PREFIX_INSTALL ?= /usr/local/bin


# tools
GO ?= $(shell which go)

APP_NAME ?= access

######
ROOT_DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
PKG_DIR := $(ROOT_DIR)${PKG_DIR_NAME}/
PROTO_DIR := $(PKG_DIR)$(PROTO_DIR_NAME)
BIN_DIR := $(ROOT_DIR)${BIN_DIR_NAME}
CMD_DIR := $(ROOT_DIR)$(CMD_DIR_NAME)
SCRIPT_DIR := $(ROOT_DIR)$(SCRIPT_DIR_NAME)
ENV_FILE := $(ROOT_DIR).env

# Project specific
SERVER_EXEC_NAME := server
PROTO_FILE := $(APP_NAME).proto
PROTO_OUT_DIR := $(PKG_DIR)

#proto
PROTO_OUT := $(PROTO_OUT_DIR)internal/
PROTO_SRC := $(PROTO_DIR)/$(PROTO_FILE)

# Version info
#VERSION  ?=  $(shell $(BASH) ./version.sh)
#GIT_COMMIT ?= $(shell $(GIT) rev-parse HEAD)
#GIT_BRANCH ?= $(shell $(GIT) rev-parse --abbrev-ref HEAD)
#GIT_REPO ?= $(shell $(GIT) config --get remote.origin.url)
BUILD_TIMESTAMP ?= $(shell $(DATE) "+%a %b %d %T %Z %Y")

# Docker
DOCKER_IMAGE_NAME := 100mslive/$(APP_NAME)
DOCKER_REPOS :=  gcr.io/dev-in-309805
DEPLOYMENT_NAME ?= $(APP_NAME)
DOCKER_GIT_USERNAME ?= s-dvd
DOCKER_GIT_TOKEN ?= ""
DOCKER_ARGS= --build-arg DOCKER_GIT_USERNAME=$(DOCKER_GIT_USERNAME) --build-arg DOCKER_GIT_TOKEN=$(DOCKER_GIT_TOKEN)
DOCKER_GIT_USERNAME := s-dvd
DOCKER_GIT_TOKEN ?= ""
DOCKER_TAG_SUFFIX ?=
DOCKER_ARGS ?= --build-arg DOCKER_GIT_USERNAME="$(DOCKER_GIT_USERNAME)"
DOCKER_ARGS +=  --build-arg DOCKER_GIT_TOKEN="$(DOCKER_GIT_TOKEN)"
DOCKER_ARGS +=  --build-arg VERSION="$(VERSION)"
DOCKER_ARGS +=  --build-arg GIT_BRANCH="$(GIT_BRANCH)"
DOCKER_ARGS +=  --build-arg GIT_REPO="$(GIT_REPO)"
DOCKER_ARGS +=  --build-arg GIT_COMMIT="$(GIT_COMMIT)"
DOCKER_TAG := $(VERSION)$(DOCKER_TAG_SUFFIX)

# Docker Compose
DOCKER_COMPOSE_PROJECT ?= hmsapi
DOCKER_COMPOSE_ARGS ?= -p $(DOCKER_COMPOSE_PROJECT)
DOCKER_COMPOSE_VARS ?= DOCKER_GIT_USERNAME="$(DOCKER_GIT_USERNAME)"
DOCKER_COMPOSE_VARS +=  DOCKER_GIT_TOKEN="$(DOCKER_GIT_TOKEN)"
DOCKER_COMPOSE_VARS +=  VERSION="$(VERSION)"
DOCKER_COMPOSE_VARS +=  GIT_BRANCH="$(GIT_BRANCH)"
DOCKER_COMPOSE_VARS +=  GIT_REPO="$(GIT_REPO)"
DOCKER_COMPOSE_VARS +=  GIT_COMMIT="$(GIT_COMMIT)"

# Build Params
BUILD_PARAMS += -X 'github.com/100mslive/packages/version.BuildTimestamp=$(BUILD_TIMESTAMP)'
BUILD_PARAMS += -X 'github.com/100mslive/packages/version.GitCommit=$(GIT_COMMIT)'
BUILD_PARAMS += -X 'github.com/100mslive/packages/version.GitBranch=$(GIT_BRANCH)'
BUILD_PARAMS += -X 'github.com/100mslive/packages/version.GitRepo=$(GIT_REPO)'
BUILD_PARAMS += -X 'github.com/100mslive/packages/version.VersionInfo=$(VERSION)'



# Linker Params
GO_LDFLAGS += -ldflags=" $(BUILD_PARAMS) "

BUILD_ENV := CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH)

BUILD_CMD := $(BUILD_ENV) $(GO) build  -installsuffix cgo  $(GO_LDFLAGS)


all: $(BIN_DIR)/$(SERVER_EXEC_NAME)
PROTO_FLAG := -I $(PROTO_OUT_DIR)internal -I $(PROTO_DIR)
PROTO_FLAG += --include_imports --descriptor_set_out=$(PROTO_SRC:.proto=.protoset)
PROTO_FLAG += --experimental_allow_proto3_optional --go_out=paths=source_relative:$(PROTO_OUT_DIR)internal
proto: $(PROTO_DIR)/$(PROTO_FILE)
	$(CLANG) -style=google -i $(PROTO_SRC)
	$(PROTOC)  $(PROTO_FLAG) $(PROTO_SRC) --go-grpc_out=$(PROTO_OUT_DIR)

.PHONY: proto
# Build
$(SERVER_EXEC_NAME): $(BIN_DIR)/$(SERVER_EXEC_NAME)
build: $(BIN_DIR)/$(SERVER_EXEC_NAME)
deps:
	$(GO) mod download

$(BIN_DIR)/$(SERVER_EXEC_NAME):
	$(BUILD_ENV) $(GO) build  -installsuffix cgo  $(GO_LDFLAGS) -o $@ $(CMD_DIR)/main.go


install:  $(BIN_DIR)/$(SERVER_EXEC_NAME)
	$(COPY) $(BIN_DIR)/$(SERVER_EXEC_NAME)  $(PREFIX_INSTALL)

.PHONY: $(BIN_DIR)/$(EXEC_NAME) install deps

# Local Debug
$(ENV_FILE):
	$(SCRIPT_DIR)/env.sh 	$@

debug:
	-$(DOCKER_COMPOSE_VARS) $(DOCKER_COMPOSE) $(DOCKER_COMPOSE_ARGS)  stop $(APP_NAME)
	$(SCRIPT_DIR)/debug.sh

.PHONY: debug


# Docker compose
deploy: $(ENV_FILE)
	$(DOCKER_COMPOSE_VARS) $(DOCKER_COMPOSE)  $(DOCKER_COMPOSE_ARGS)  up --build -d --remove-orphans

logs:
	$(DOCKER_COMPOSE_VARS) $(DOCKER_COMPOSE)  $(DOCKER_COMPOSE_ARGS) logs -f api


.PHONY:  deploy logs

# docker
docker:
	$(DOCKER) build -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) $(DOCKER_ARGS) .

release: docker
	$(SCRIPT_DIR)/release.sh --image $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) --name $(DEPLOYMENT_NAME) --tag $(DOCKER_TAG) $(DOCKER_REPOS)

.PHONY: docker docker-release release

# Cleanup
clean:
	- $(DOCKER_COMPOSE)  -p $(DOCKER_COMPOSE_PROJECT) down
	$(RM) $(BIN_DIR)/$(EXEC_NAME)

cleanall: clean
	- $(DOCKER) volume rm $(DOCKER_COMPOSE_PROJECT)_mongodb.data

.PHONY:  clean cleanall

