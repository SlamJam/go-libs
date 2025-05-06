test-all:
    @go test -race -timeout 5m -count=1 -coverprofile=coverage.out ./...
    @go tool cover -html=coverage.out -o coverage.html


# go test -race -timeout 5m -count=1 -coverprofile=coverage.out -covermode=atomic -v -bench=. -benchmem ./...
# go test -race -timeout 5m ./... -tags=goleak
