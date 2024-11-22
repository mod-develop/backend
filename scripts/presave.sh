gofmt ./internal/.. ./pkg/.. ./cmd/..
goimports -local "github.com/mod-develop/backend" -w ./internal/.. ./pkg/.. ./cmd/..
go mod tidy
go test ./...