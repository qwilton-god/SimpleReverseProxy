# Reverse Proxy

Простой reverse proxy с балансировкой нагрузки, написанный на Go с нуля.

## Особенности

- Round-robin балансировка между серверами
- Health check для проверки доступности бэкендов
- Ручное проксирование HTTP запросов
- Graceful shutdown
- Thread-safe операции

## Структура

```
proxy/          - Reverse proxy с балансировкой
├── backend.go       - Структура бэкенда
├── pool.go          - Пул серверов (round-robin)
├── healthcheck.go   - Проверка доступности
├── proxy.go         - Ручное проксирование запросов
├── loadbalancer.go  - HTTP handler
└── main.go          - Точка входа

server/         - Тестовые сервера
└── main.go          - Простой HTTP сервер
```

## Запуск

```bash
# Сборка
make build

# Запуск прокси
./bin/proxy

# Запуск тестовых серверов
./bin/server :8081
./bin/server :8082
./bin/server :8083
```

## Тестирование

```bash
curl http://localhost:8080
```

## Конфигурация

Переменные окружения:
- `PROXY_PORT` - порт прокси (по умолчанию `:8080`)
- `SERVER_1_PORT`, `SERVER_2_PORT`, `SERVER_3_PORT` - порты бэкендов
- `HEALTH_CHECK_INTERVAL` - интервал проверки в секундах (по умолчанию `10`)
