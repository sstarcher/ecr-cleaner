TAG_NAME := $(shell git tag -l --contains HEAD)
IMAGE_REPOSITORY_NAME=sstarcher/ecr-cleaner:$(TAG_NAME)

DOCKER_BUILD_PLATFORMS ?= linux/amd64,linux/arm64

build:
	docker buildx build $(DOCKER_BUILDX_ARGS) --progress=chain -t $(IMAGE_REPOSITORY_NAME) --platform=$(DOCKER_BUILD_PLATFORMS) -f Dockerfile .

.PHONY: build

