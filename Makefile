APP := holidaychecker

.PHONY: tidy test build run
tidy: ; go mod tidy
test: ; go test ./... -race -v
build: ; CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(APP) ./cmd/holidaychecker
run: ; go run ./cmd/holidaychecker -date=2025/01/01 -countries=ES,FR,IT