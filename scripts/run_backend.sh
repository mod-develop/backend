swag init -o ./docs -g ./internal/adapters/api/rest/rest.go
swag fmt
go run ./cmd/server/server.go