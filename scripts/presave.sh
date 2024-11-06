gofmt ./internal/.. ./pkg/.. ./cmd/..
goimports -local "github.com/playmixer/medal-of-discipline" -w ./internal/.. ./pkg/.. ./cmd/..
go mod tidy
go test ./...