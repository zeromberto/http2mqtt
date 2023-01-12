all: verify build
verify: lint
prepare: fmt verify

.PHONY: all verify prepare generate fmt lint build run


fmt: $(GOIMPORTS)
	$(call go_get,goimports,golang.org/x/tools/cmd/goimports)
	@goimports -w .
	@go mod tidy

lint:
	@golangci-lint run --skip-files internal/registrations/gateway.go

build:
	@CGO_ENABLED=0 go build -trimpath -o ./artifacts/service ./service.go

run: build
	@./artifacts/service
