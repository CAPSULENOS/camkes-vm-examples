all: build/aegis

SOURCES := $(shell find . -name '*.go')

build/aegis: ${SOURCES}
	mkdir -p build
	go build -o $@ cmd/aegis/main.go

.PHONY: clean
clean:
	@rm -rf build
