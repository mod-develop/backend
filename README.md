# MOD

### start start
```bash
docker-compose up
```

### rebuild
```bash
docker-compose up --build
```

### down server
```bash
docker-compose down
```

### env example
```env
HTTP_ADDRESS=localhost:8080
SECRET_KEY=secret_key

DATABASE_URI=postgres://root:root@localhost:5432/discipline?sslmode=disable&TimeZone=Europe/Moscow

LOG_LEVEL=debug
```