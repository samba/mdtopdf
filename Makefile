
all: test build install

test:
	go test ./...

build:
	go build ./cmd/md2pdf

install:
	go install ./cmd/md2pdf
