package rest

type Config struct {
	Address   string `env:"HTTP_ADDRESS"`
	SecretKey string `env:"SECRET_KEY"`
}
