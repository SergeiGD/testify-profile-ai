# Testify Profile

> Этот проект создан **полностью с помощью vibe coding** — от архитектуры до реализации и написания тест-кейсов, с использованием AI-ассистента на каждом этапе разработки.

Сервис управления профилями пользователей для проекта Testify. Реализует сервис управления профилями.

## Стек технологий

| Категория       | Технология                                             |
|-----------------|--------------------------------------------------------|
| Язык            | Go 1.25                                                |
| HTTP фреймворк  | [Gin](https://github.com/gin-gonic/gin)                |
| gRPC            | [google.golang.org/grpc](https://github.com/grpc/grpc-go) + Protobuf |
| База данных      | PostgreSQL (драйвер [pgx v5](https://github.com/jackc/pgx)) |
| Миграции        | [Goose](https://github.com/pressly/goose)              |
| Кодогенерация   | [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) (OpenAPI → Go), `protoc` (Protobuf → Go) |
| DI              | [Google Wire](https://github.com/google/wire)          |
| Моки            | [Mockery](https://github.com/vektra/mockery)           |
| Тесты           | [Testify](https://github.com/stretchr/testify)         |
| Конфигурация    | [cleanenv](https://github.com/ilyakaznacheev/cleanenv) |
| Логирование     | [Logrus](https://github.com/sirupsen/logrus)           |
| Контейнеризация | Docker / Docker Compose                                |
| API документация| Swagger UI                                             |

## Архитектура

Проект следует принципам **Clean Architecture** с чётким разделением слоёв:

```
internal/
├── api/
│   ├── http/           # HTTP-обработчики (сгенерированы из OpenAPI + кастомные хендлеры)
│   └── grpc/           # gRPC-обработчики (сгенерированы из Protobuf)
│       └── pb/         # Сгенерированный Protobuf-код
├── server/
│   ├── httpserv/       # Запуск HTTP-сервера (Gin)
│   └── grpcserv/       # Запуск gRPC-сервера (с API-key интерцептором)
├── usecases/
│   ├── register/       # Бизнес-логика регистрации
│   ├── auth/           # Бизнес-логика аутентификации (login, refresh)
│   └── profile/        # Бизнес-логика получения профиля
├── adapter/
│   ├── postgres/       # Репозитории (PostgreSQL)
│   └── smtp/           # Email-отправитель
├── services/
│   ├── clock/          # Абстракция времени (удобно для тестов)
│   ├── token/          # Генерация токенов подтверждения
│   ├── password/       # Хэширование паролей
│   ├── jwt/            # Генерация и верификация JWT (RS256)
│   └── linkbuilder/    # Построение ссылок подтверждения
├── domain/             # Доменные ошибки
├── models/             # Модели данных (User, RegistrationConfirmation)
└── dto/                # DTO для юзкейсов
```

Оба сервера (HTTP и gRPC) запускаются параллельно через `errgroup` и завершаются по одному контексту при получении `SIGINT`/`SIGTERM`.

## API

### HTTP REST

Документация доступна через Swagger UI после запуска: [http://localhost:8081](http://localhost:8081)

Спецификация: [`docs/oapi/api.yaml`](docs/oapi/api.yaml)

| Метод  | Путь                                | Описание                                    |
|--------|-------------------------------------|---------------------------------------------|
| GET    | `/api/v1/healthcheck/`              | Проверка работоспособности сервиса          |
| POST   | `/api/v1/register/`                 | Шаг 1: создание аккаунта, отправка письма   |
| GET    | `/api/v1/register/confim/{token}`   | Шаг 2: подтверждение email по токену        |
| POST   | `/api/v1/auth/login`                | Вход: получение access + refresh токенов    |
| POST   | `/api/v1/auth/refresh`              | Обновление пары токенов по refresh токену   |

#### Пример запроса регистрации

```json
POST /api/v1/register/
{
  "email": "user@example.com",
  "password": "secret123",
  "username": "john_doe",
  "birth_date": "1995-06-15"
}
```

### gRPC

Сервис доступен на порту `50051` (настраивается через `GRPC_PORT`).

Спецификация: [`api/proto/profile/profile.proto`](api/proto/profile/profile.proto)

| Метод                           | Описание                                          |
|---------------------------------|---------------------------------------------------|
| `ProfileService/GetProfile`     | Получение профиля пользователя по UUID            |

Все вызовы требуют передачи API-ключа в заголовке метаданных `x-api-key` (значение задаётся в `GRPC_API_KEY`). При отсутствии или неверном ключе возвращается статус `UNAUTHENTICATED`.

#### Пример ответа GetProfile

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "username": "john_doe",
  "birth_date": "1995-06-15"
}
```

## База данных

Миграции управляются через Goose и запускаются автоматически при старте через Docker Compose.

| Таблица                      | Описание                                      |
|------------------------------|-----------------------------------------------|
| `users`                      | Пользователи (UUID PK, email, password_hash, username, birth_date, is_confirmed) |
| `registration_confirmations` | Токены подтверждения регистрации с TTL        |

## Запуск

### Предварительные требования

- Docker и Docker Compose

### Настройка окружения

Скопируйте файл-пример и заполните нужные значения:

```bash
cp .env.example .env
```

`.env.example` содержит все необходимые переменные:

| Переменная                    | Описание                                              |
|-------------------------------|-------------------------------------------------------|
| `DATABASE_HOST`               | Хост PostgreSQL                                       |
| `DATABASE_USER`               | Пользователь БД                                       |
| `DATABASE_PASSWORD`           | Пароль БД                                             |
| `DATABASE_NAME`               | Имя БД                                                |
| `DATABASE_PORT`               | Порт БД                                               |
| `DATABASE_MAXATTEMPS`         | Максимальное число попыток подключения к БД           |
| `DATABASE_DELAY`              | Задержка между попытками подключения                  |
| `DATABASE_TIMEOUT`            | Таймаут подключения к БД                              |
| `GOOSE_DRIVER`                | Драйвер для Goose (`postgres`)                        |
| `GOOSE_DBSTRING`              | DSN строка для Goose (применение миграций)            |
| `DEBUG`                       | Режим отладки (`true` / `false`)                      |
| `SMTP_HOST`                   | Хост SMTP-сервера                                     |
| `SMTP_PORT`                   | Порт SMTP-сервера                                     |
| `SMTP_USERNAME`               | Логин SMTP                                            |
| `SMTP_PASSWORD`               | Пароль SMTP                                           |
| `SMTP_FROM`                   | Адрес отправителя писем                               |
| `APP_BASE_URL`                | Базовый URL приложения (используется в письмах)       |
| `REGISTRATION_TOKEN_TTL`      | Время жизни токена подтверждения регистрации          |
| `REGISTRATION_RESEND_COOLDOWN`| Минимальный интервал между повторными отправками      |
| `USE_MOCK_EMAIL_SENDER`       | Использовать мок вместо реального SMTP (`true`/`false`)|
| `JWT_PRIVATE_KEY`             | RSA приватный ключ (PEM) для подписи JWT (RS256)       |
| `JWT_PUBLIC_KEY`              | RSA публичный ключ (PEM) для верификации JWT           |
| `JWT_ACCESS_TTL`              | Время жизни access токена (например, `15m`)            |
| `JWT_REFRESH_TTL`             | Время жизни refresh токена (например, `168h`)          |
| `GRPC_PORT`                   | Порт gRPC-сервера (по умолчанию `50051`)              |
| `GRPC_API_KEY`                | API-ключ для аутентификации gRPC-запросов             |

### Запуск через Docker Compose

```bash
make run
```

Это запустит:
- **PostgreSQL** — база данных (порт из `.env`)
- **migrate** — автоматическое применение миграций
- **app** — Go-сервер: HTTP на порту `8080`, gRPC на порту `50051`
- **swagger-ui** — документация API на порту `8081`

## Разработка

### Полезные команды

```bash
# Привести зависимости в порядок
make tidy

# Сгенерировать HTTP-обработчики из OpenAPI-спецификации
make codegen-api

# Сгенерировать моки для тестов
make codegen-mocks

# Запустить тесты
make tests
```

### Регенерация Protobuf

gRPC-код генерируется из [`api/proto/profile/profile.proto`](api/proto/profile/profile.proto). Для перегенерации:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/profile/profile.proto
```

Сгенерированные файлы сохраняются в `internal/api/grpc/pb/`.
