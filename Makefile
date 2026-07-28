.PHONY: help
help:
	@go run ./cmd/main.go -help

.PHONY: list
list:
	@go run ./cmd/main.go -list

.PHONY: add
add:
	@go run ./cmd/main.go -add

.PHONY: complete
complete:
	@go run ./cmd/main.go -complete

.PHONY: test
test:
	@go test ./... -cover
