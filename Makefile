ifeq ($(OS), Windows_NT)
	VERSION := $(shell git describe --exact-match --tags 2>nil)
else
	VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
endif

COMMIT ?= $(shell git rev-parse --short=8 HEAD)

BINARY=pigeon
TARGET = $(word 1,$(subst -, ,$*))

build:
	go build -o bin/${BINARY} ${LDFLAGS} ./cmd/pigeon/*.go