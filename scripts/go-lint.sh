#!/bin/bash

# Устанавливаем рабочую директорию скрипта как текущую директорию
# cd "$(dirname "$0")"

# Определяем пути к директориям с исходным кодом и линтером
source_dir="$(pwd)"
golint_cache_dir="$source_dir/golangci-lint/.cache/golangci-lint/v1.61.0"

echo $source_dir
echo $golint_cache_dir

# Запускаем Docker с нужными параметрами
docker run --rm \
            -v "$source_dir:/app" \
            -v "$golint_cache_dir:/root/.cache" \
            -w /app golangci/golangci-lint:v1.61.0 \
            golangci-lint run -c .golangci.yml > ./golangci-lint/report-unformatted.json