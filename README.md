# IDP — Identity Provider

Небольшой gRPC сервис аутентификации для моих проектов

## Возможности

- Регистрация пользователей и логин (access token и refresh token)
- Ротация refresh-токенов по семействам
- Private Key JWT аутентификация клиентов
- Периодическая ротация ключей подписи JWT, публикация JWKS

## Требования

| Компонент  | Версия  |
| ---------- | ------- |
| Go         | >= 1.27 |
| PostgreSQL | >= 18.6 |
| Redis      | >= 8.10 |
| goose      | >= 3.27 |
| sqlc       | >= 1.31 |
| task       | >= 3.45 |
| protoc     | >= 33.4 |

## Структура проекта

```
idp/
├── cmd/
│   ├── idp/                  # Точка входа основного сервиса
│   └── jwks-mock/            # Мок внешнего клиента 
├── db/
│   └── sql/
│       ├── migrations/       # goose-миграции
│       ├── queries/          # SQL-запросы для sqlc
│       └── gen/dbstore/      # Сгенерированный код sqlc
├── internal/
│   ├── app/                  # Composition root
│   │   ├── app.go
│   │   ├── init.go
│   │   ├── http.go
│   │   ├── grpc.go
│   │   ├── workers.go
│   │   └── grpc/interceptor/
│   ├── domain/               # Бизнес-логика
│   │   ├── model/
│   │   ├── user/
│   │   ├── client/
│   │   ├── auth/
│   │   └── error.go          # DomainError и коды ошибок
│   └── lib/                  # Инфраструктурные утилиты
│       ├── config/
│       ├── crypt/            # Криптографические утилиты
│       │   ├── hash/         # Хэширование
│       │   │   ├── argon2/   # Алгоритм Argon2
│       │   │   └── blake3/   # Алгоритм Blake3
│       │   └── keys/         # Пакет управления ключами
│       ├── valid/            # Утилиты валидации
│       ├── sign/             # Утилита именования операций/ошибок
│       └── meta/             # Ключи context
└── proto/                    # Protobuf контракты
    ├── idp/idp.proto         # Контракты сервиса, standalone репозиторий
    └── gen/idp/              # Сгенерированный код protobuf

```

## API

### gRPC API

#### `AuthService`

| RPC            | Назначение                                         |
| -------------- | -------------------------------------------------- |
| `Register`     | Регистрация пользователя, возвращает `user_id`     |
| `Login`        | Логин, возвращает `access_token` + `refresh_token` |
| `RefreshToken` | Обновление пары токенов (ротация refresh)          |
| `Logout`       | Отзыв refresh-токена                               |

#### `ClientService`

| RPC            | Назначение                                                |
| -------------- | --------------------------------------------------------- |
| `ClientCreate` | Регистрация внешнего клиента (публичный RPC)              |
| `ClientDelete` | Удаление клиента (id берётся из контекста аутентификации) |

### HTTP API

| HTTP                    | Назначение                   |
| ----------------------- | ---------------------------- |
| `GET /.well-known/jwks` | Получение JWKS (ECDSA P-256) |

## ENV

### Обязательные

| Переменная           | Назначение                          |
| -------------------- | ----------------------------------- |
| `SERVICE_NAME`       | Имя сервиса (issuer JWT)            |
| `ENV`                | Окружение: `local` / `dev` / `prod` |
| `DB_CONN_STR`        | Строка подключения PostgreSQL       |
| `REDIS_CONN_STR`     | Строка подключения Redis            |
| `HTTP_ADDR`          | Адрес HTTP-сервера           |
| `GRPC_ADDR`          | Адрес gRPC-сервера                  |

### Опциональные (дефолты)

| Переменная                        | Дефолт    | Назначение                               |
| --------------------------------- | --------- | ---------------------------------------- |
| `ACCESS_TOKEN_TTL`                | `15m`     | Время жизни access-токена                |
| `REFRESH_TOKEN_TTL`               | `720h`    | Время жизни refresh-токена               |
| `REFRESH_TOKEN_REVOKED_STORE_TTL` | `10m`     | TTL хранения отозванных refresh-токенов  |
| `REFRESH_TOKEN_GRACE_PERIOD`      | `30s`     | Grace period при ротации refresh-токенов |
| `GRPC_TIMEOUT`                    | `5s`      | Таймаут gRPC                             |
| `KEY_ROTATION_INTERVAL`           | `24h`     | Интервал ротации ключей          |
| `KEY_ROTATION_GRACE_PERIOD`       | `1h`      | Grace period ротации ключей              |
| `HASH_MEMORY`                     | `2097152` | Argon2: память (байт)                    |
| `HASH_ITERATIONS`                 | `1`       | Argon2: итерации                         |
| `HASH_PARALLELISM`                | `NumCPU`  | Argon2: параллелизм                      |
| `HASH_KEY_LENGTH`                 | `32`      | Argon2: длина ключа                      |
| `HASH_SALT_LENGTH`                | `16`      | Argon2: длина соли                       |

## Как запустить

### Шаги

1. **Подготовить окружение:**

   ```bash
   cp .env.example .env.local
   ```

   Заполнить `DB_CONN_STR`, `REDIS_CONN_STR`.

2. **Применить миграции:**

   ```bash
   task migration-up
   ```

3. **Запустить IdP:**

   ```bash
   task start
   ```

   gRPC-сервер — на `GRPC_ADDR` (по умолчанию `localhost:50051`), HTTP (JWKS) — на `HTTP_ADDR` (по умолчанию `localhost:50050`).

4. **(Опционально) Запустить JWKS mock** — имитация внешнего клиента:

   ```bash
   task tjwks <clientId-uuid>
   ```

   Поднимает HTTP-сервер на `:8081`:
   - `GET /.well-known/jwks` — JWKS (RSA-2048, RS256)
   - `GET /bearer` — подписанный JWT клиента в формате `{"Bearer": base64("clientId:token")}`
