# Gift Buyer - автоматическая покупка Telegram подарков

[![Language: Russian](https://img.shields.io/badge/Language-Русский-blue)](#русский) [![Language: English](https://img.shields.io/badge/Language-English-green)](#english) [![Telegram](https://img.shields.io/badge/Telegram-@chiefssq-blue?logo=telegram)](https://t.me/cheifssq)

## 📑 Содержание

### Русский
- [📖 Описание](#-описание)
- [⚡ Преимущества](#-преимущества)
- [🚀 Быстрый старт](#-быстрый-старт)
  - [Установка зависимостей](#установка-зависимостей)
  - [Настройка переменных окружения](#настройка-переменных-окружения)
  - [Сборка и запуск](#сборка-и-запуск)
- [🛠 Разработка](#-разработка)
  - [Доступные команды](#доступные-команды)
  - [Локальная проверка перед коммитом](#локальная-проверка-перед-коммитом)
- [📁 Структура проекта](#-структура-проекта)
- [🔒 Безопасность](#-безопасность)
- [📊 Тестирование](#-тестирование)
- [🚀 Деплой](#-деплой)
- [📝 Логирование](#-логирование)
- [🤝 Вклад в проект](#-вклад-в-проект)
- [📄 Лицензия](#-лицензия)
- [⚙️ Подробная конфигурация](#️-подробная-конфигурация)
  - [🔧 Telegram настройки](#-telegram-настройки)
  - [🎯 Критерии покупки](#-критерии-покупки)
  - [👤 Получатели подарков](#-получатели-подарков)
  - [⏱️ Частота проверки](#️-частота-проверки)
  - [🚀 Параметры производительности](#-параметры-производительности)
  - [🧪 Тестовый режим](#-тестовый-режим)
  - [🔒 Глобальные ограничители](#-глобальные-ограничители)
- [📋 Полный пример конфигурации](#-полный-пример-конфигурации)

### English
- [📖 Description](#-description)
- [⚡ Advantages](#-advantages)
- [🚀 Quick Start](#-quick-start-1)
- [🛠 Development](#-development)
- [📁 Project Structure](#-project-structure)
- [🔒 Security](#-security)
- [📊 Testing](#-testing)
- [🚀 Deploy](#-deploy)
- [📝 Logging](#-logging)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)
- [⚙️ Detailed Configuration](#️-detailed-configuration)

---

## Русский

### 📖 Описание

**Gift Buyer** — софт для покупки Gifts в Telegram. Программа непрерывно мониторит доступные подарки, проверяет их соответствие заданным критериям и автоматически покупает подходящие варианты.

### ⚡ Преимущества

- **Высокая скорость** — настраиваемый тикер для мониторинга и мгновенной реакции на появление новых подарков
- **Точная фильтрация** — настраиваемые критерии по цене, количеству и лимитам
- **Параллельная обработка** — одновременная покупка нескольких подарков
- **Оптимизированная производительность** — RPC лимит 50 запросов/сек для максимальной скорости, параметр настраиваемый
- **Уведомления** — мгновенные оповещения в Telegram при необходимости
- **Кэширование** — сохранение состояния между перезапусками
- **Безопасность** — graceful shutdown и обработка ошибок

### 🚀 Быстрый старт

### Установка зависимостей
```bash
go mod download
```

### Настройка переменных окружения
### Сборка и запуск
```bash
go build -o gift-buyer cmd/main.go
./gift-buyer
```

## 🛠 Разработка

### Доступные команды
```bash
go mod tidy              # Обновить зависимости
go test ./...            # Запустить тесты
go test -v ./...         # Запустить тесты с подробным выводом
go test -cover ./...     # Запустить тесты с покрытием
go vet ./...             # Проверить код статическим анализатором
go fmt ./...             # Отформатировать код
go build -o gift-buyer cmd/main.go  # Собрать бинарный файл
go run cmd/main.go       # Запустить без сборки
```

### Локальная проверка перед коммитом
```bash
go test ./... && go vet ./... && go fmt ./...
```

## 📁 Структура проекта

```
├── cmd/                    # Точки входа приложения
├── internal/              # Внутренняя логика
│   ├── config/           # Конфигурация
│   └── service/          # Бизнес-логика
├── pkg/                  # Переиспользуемые пакеты
├── .github/workflows/    # GitHub Actions
├── docs/                 # Документация
├── go.mod               # Зависимости Go
├── .golangci.yml        # Конфигурация линтера
└── env.example          # Пример переменных окружения
```

## 🔒 Безопасность

Проект использует:
- Криптографически стойкие генераторы случайных чисел
- Безопасные права доступа к файлам (0600)
- Проверку безопасности с помощью gosec
- Обработку всех ошибок

## 📊 Тестирование

```bash
go test ./...           # Запустить все тесты
go test -cover ./...    # Тесты с отчетом о покрытии
```

## 🚀 Деплой

### Ручная сборка
```bash
# Сборка для всех платформ
GOOS=linux GOARCH=amd64 go build -o gift-buyer-linux cmd/main.go
GOOS=windows GOARCH=amd64 go build -o gift-buyer-windows.exe cmd/main.go
GOOS=darwin GOARCH=amd64 go build -o gift-buyer-macos cmd/main.go
```

## 📝 Логирование

Уровни логирования настраиваются через переменную `LOG_LEVEL`:
- `debug` - Подробная отладочная информация
- `info` - Общая информация (по умолчанию)
- `warn` - Предупреждения
- `error` - Только ошибки

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для фичи (`git checkout -b feature/amazing-feature`)
3. Запустите `go test ./... && go vet ./...` для проверки
4. Закоммитьте изменения (`git commit -m 'Add amazing feature'`)
5. Запушьте ветку (`git push origin feature/amazing-feature`)
6. Создайте Pull Request

## 📄 Лицензия

Этот проект распространяется под лицензией MIT.

### 🚀 Быстрый запуск

1. **Скачайте и соберите проект:**
   ```bash
   git clone <repository-url>
   cd gift-buyer
   go build -o gift-buyer cmd/main.go
   ```

2. **Настройте конфигурацию:**
   ```bash
   cp internal/config/config_example.json internal/config/config.json
   # Отредактируйте config.json с вашими данными
   ```

3. **Запустите программу:**
   ```bash
   ./gift-buyer
   ```

### ⚙️ Подробная конфигурация

#### 🔧 Telegram настройки (`tg_settings`)

```json
{
    "tg_settings": {
        "app_id": 12345678,
        "api_hash": "ваш_api_hash",
        "phone": "+1234567890",
        "password": "пароль_2fa",
        "tg_bot_key": "токен_бота",
        "notification_chat_id": 123456789
    }
}
```

- **`app_id`** и **`api_hash`** — обязательные параметры из [my.telegram.org](https://my.telegram.org)
- **`phone`** — номер телефона аккаунта в международном формате
- **`password`** — пароль двухфакторной аутентификации (можно оставить пустым `""` если 2FA отключена)
- **`tg_bot_key`** — токен Telegram бота для уведомлений (можно оставить пустым `""` если уведомления не нужны)
- **`notification_chat_id`** — ID чата для отправки уведомлений(ваш юзер айди)

#### 🎯 Критерии покупки (`criterias`)

Можно указать несколько критериев - программа будет покупать подарки, соответствующие любому из них:

```json
{
    "criterias": [
        {
            "min_price": 10,
            "max_price": 100,
            "total_supply": 1000,
            "count": 2
        },
        {
            "min_price": 500,
            "max_price": 1000,
            "total_supply": 100,
            "count": 1
        }
    ]
}
```

- **`min_price`** — минимальная цена подарка в звездах
- **`max_price`** — максимальная цена подарка в звездах
- **`total_supply`** — максимальный общий тираж подарка
- **`count`** — количество подарков для покупки по этому критерию

#### 👤 Получатели подарков (`receiver`)

```json
{
    "receiver": {
        "type": [0, 1, 2],
        "user_receiver_id": [123456789, 987654321],
        "channel_receiver_id": [-1001234567890]
    }
}
```

- **`type`** — массив типов получателей:
  - `0` — отправить себе
  - `1` — отправить другому пользователю (пользователь должен быть в ваших контактах)
  - `2` — отправить в канал/супергруппу (канал должен принадлежать вашему аккаунту)
- **`user_receiver_id`** — массив ID пользователей для получения подарков
- **`channel_receiver_id`** — массив ID каналов/супергрупп для получения подарков (используйте полный формат, например `-1001234567890`)

**Как это работает:**
- При каждой покупке система случайно выбирает один тип из массива `type`
- В зависимости от типа выбирается случайный ID из соответствующего массива (`user_receiver_id` или `channel_receiver_id`)
- Это позволяет распределять подарки между несколькими получателями разных типов
- Если указать `type: [0]`, все подарки будут отправляться себе (ID игнорируются)

#### ⏱️ Частота проверки (`ticker`)

```json
{
    "ticker": 2.0
}
```

Интервал между проверками новых подарков в секундах (по умолчанию 2.0 секунды)

#### 🚀 Параметры производительности

Новые параметры для оптимизации производительности системы:

```json
{
    "retry_count": 3,
    "retry_delay": 0.5,
    "concurrency_gift_count": 10,
    "concurrent_operations": 300,
    "rpc_rate_limit": 50
}
```

- **`retry_count`** — количество попыток повтора при неудачной покупке (по умолчанию 3)
- **`retry_delay`** — задержка между попытками повтора в секундах (по умолчанию 0.5 секунды)
- **`concurrency_gift_count`** — максимальное количество подарков, обрабатываемых одновременно (по умолчанию 10)
- **`concurrent_operations`** — максимальное количество одновременных операций (по умолчанию 300)
- **`rpc_rate_limit`** — лимит RPC запросов в секунду для Telegram API (по умолчанию 50 RPS)

**Оптимизация производительности:**
- Система использует асинхронную обработку повторов
- RPC запросы ограничены до 30 в секунду для соблюдения лимитов Telegram
- Параллельная обработка до 300 операций одновременно
- Минимальная задержка между повторами (10мс) для максимальной скорости

#### 🧪 Тестовый режим (`test_mode`)

```json
{
    "test_mode": true
}
```

В тестовом режиме:
- **Не учитывается** общий тираж (`total_supply`)
- **Не учитывается** капитализация
- **Поле лимитированности должно быть отрицательным** для покупки нелимитированных подарков
- Подарки покупаются без реальных ограничений

#### 🔒 Глобальные ограничители

```json
{
    "total_star_cap": 10000,
    "max_buy_count": 5,
    "limited_status": false
}
```

- **`total_star_cap`** — максимальная капитализация подарка в звездах
- **`max_buy_count`** — глобальный лимит покупок
- **`limited_status`** — фильтр по статусу лимитированности подарков (true = только лимитированные, false = все подарки)

**Пример работы глобального ограничителя:**
Если у вас есть критерии на покупку 3+2+1=6 подарков, но `max_buy_count: 4`, то купится только 4 подарка (в порядке появления).

### 📋 Полный пример конфигурации

```json
{
    "logger_level": "info",
    "soft_config": {
        "tg_settings": {
            "app_id": 12345678,
            "api_hash": "ваш_api_hash",
            "phone": "+1234567890",
            "password": "",
            "tg_bot_key": "",
            "notification_chat_id": 123456789
        },
        "criterias": [
            {
                "min_price": 10,
                "max_price": 50,
                "total_supply": 1000,
                "count": 3
            },
            {
                "min_price": 100,
                "max_price": 500,
                "total_supply": 500,
                "count": 2
            }
        ],
        "total_star_cap": 5000,
        "receiver": {
            "type": [0, 1],
            "user_receiver_id": [0, 123456789, 987654321],
            "channel_receiver_id": []
        },
        "test_mode": false,
        "max_buy_count": 4,
        "ticker": 2.0,
        "retry_count": 4,
        "retry_delay": 0.01,
        "limited_status": false,
        "concurrency_gift_count": 10,
        "concurrent_operations": 300,
        "rpc_rate_limit": 30
    }
}
```

### 📋 Требования

- Go 1.23.4+
- Telegram аккаунт с API ключами(https://my.telegram.org/apps)
- Telegram бот для уведомлений (опционально)

---

## English

### 📖 Description

**Gift Buyer** is an automated system for purchasing Star Gifts in Telegram. The program continuously monitors available gifts, validates them against configured criteria, and automatically purchases eligible options.

### ⚡ Advantages

- **🚀 High Speed** — instant reaction to new gift appearances
- **🎯 Precise Filtering** — configurable criteria for price, quantity, and limits
- **⚙️ Parallel Processing** — simultaneous purchase of multiple gifts (up to 300 operations)
- **🔄 Asynchronous Retries** — intelligent retry system with minimal delays (10ms)
- **⚡ Optimized Performance** — RPC rate limit of 30 requests/sec for maximum speed
- **📱 Notifications** — instant Telegram alerts
- **💾 Caching** — state persistence between restarts
- **🛡️ Security** — graceful shutdown and error handling

### 🚀 Quick Start

### Install dependencies
```bash
go mod download
```

### Set up environment variables
Copy the example environment file:
```bash
cp env.example .env
```

Fill in the `.env` file with your data:
- `TG_APP_ID` - Telegram application ID
- `TG_API_HASH` - API Hash from Telegram
- `TG_PHONE` - Phone number
- `TG_PASSWORD` - Password (optional)

### Build and run
```bash
go build -o gift-buyer cmd/main.go
./gift-buyer
```

## 🛠 Development

### Available commands
```bash
go mod tidy              # Update dependencies
go test ./...            # Run tests
go test -v ./...         # Run tests with verbose output
go test -cover ./...     # Run tests with coverage
go vet ./...             # Check code with static analyzer
go fmt ./...             # Format code
go build -o gift-buyer cmd/main.go  # Build binary file
go run cmd/main.go       # Run without building
```

### Local check before commit
```bash
go test ./... && go vet ./... && go fmt ./...
```

## 📁 Project Structure

```
├── cmd/                    # Application entry points
├── internal/              # Internal logic
│   ├── config/           # Configuration
│   └── service/          # Business logic
├── pkg/                  # Reusable packages
├── .github/workflows/    # GitHub Actions
├── docs/                 # Documentation
├── go.mod               # Go dependencies
├── .golangci.yml        # Linter configuration
└── env.example          # Environment variables example
```

## 🔒 Security

The project uses:
- Cryptographically secure random number generators
- Secure file permissions (0600)
- Security checks with gosec
- Handling of all errors

## 📊 Testing

```bash
go test ./...           # Run all tests
go test -cover ./...    # Tests with coverage report
```

Coverage report is saved to `coverage.html`.

## 🚀 Deploy

### Automatic deploy
When pushing to `main` branch automatically:
1. All tests are run
2. Binary files are built for all platforms
3. Release is created with artifacts

### Manual build
```bash
# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o gift-buyer-linux cmd/main.go
GOOS=windows GOARCH=amd64 go build -o gift-buyer-windows.exe cmd/main.go
GOOS=darwin GOARCH=amd64 go build -o gift-buyer-macos cmd/main.go
```

## 📝 Logging

Logging levels are configured via `LOG_LEVEL` variable:
- `debug` - Detailed debug information
- `info` - General information (default)
- `warn` - Warnings
- `error` - Errors only

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Run `go test ./... && go vet ./...` for checks
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push the branch (`git push origin feature/amazing-feature`)
6. Create a Pull Request

## 📄 License

This project is distributed under the MIT License.



### ⚙️ Detailed Configuration

#### 🔧 Telegram Settings

**⚠️ IMPORTANT:** Telegram settings are loaded from environment variables, NOT from JSON file!

Required environment variables:
```bash
export TG_APP_ID=12345678
export TG_API_HASH=your_api_hash
export TG_PHONE=+1234567890
```

Optional environment variables:
```bash
export TG_PASSWORD=2fa_password           # Only if 2FA is enabled
export TG_BOT_KEY=bot_token              # For notifications
export TG_NOTIFICATION_CHAT_ID=123456789 # Chat ID for notifications
export DEVICE_MODEL="MacBook Pro M1 Pro" # Device model
export SYSTEM_VERSION="macOS 15.3.1"     # System version
export APP_VERSION="11.9 (272031) APP_STORE" # App version
export SYSTEM_LANG_CODE=en               # System language
export LANG_CODE=en                      # Interface language
export LANG_PACK=en                      # Language pack
```

In JSON file, the `tg_settings` section is ignored - all values are taken from environment variables.

**Parameter descriptions:**
- **`TG_APP_ID`** and **`TG_API_HASH`** — required parameters from [my.telegram.org](https://my.telegram.org)
- **`TG_PHONE`** — account phone number in international format
- **`TG_PASSWORD`** — two-factor authentication password (can be omitted if 2FA is disabled)
- **`TG_BOT_KEY`** — Telegram bot token for notifications (optional)
- **`TG_NOTIFICATION_CHAT_ID`** — chat ID for sending notifications

#### 🎯 Purchase Criteria (`criterias`)

You can specify multiple criteria - the program will purchase gifts matching any of them:

```json
{
    "criterias": [
        {
            "min_price": 10,
            "max_price": 100,
            "total_supply": 1000,
            "count": 2
        },
        {
            "min_price": 500,
            "max_price": 1000,
            "total_supply": 100,
            "count": 1
        }
    ]
}
```

- **`min_price`** — minimum gift price in stars
- **`max_price`** — maximum gift price in stars
- **`total_supply`** — maximum total gift supply
- **`count`** — number of gifts to purchase for this criteria

#### 👤 Gift Recipients (`receiver`)

```json
{
    "receiver": {
        "type": [1, 2],
        "user_receiver_id": [123456789, 987654321],
        "channel_receiver_id": [555666777, 888999000]
    }
}
```

- **`type`** — array of recipient types:
  - `0` — send to yourself
  - `1` — send to another user (user must be in your contacts)
  - `2` — send to channel (channel must belong to your account)
- **`user_receiver_id`** — array of user IDs for gift recipients
- **`channel_receiver_id`** — array of channel IDs for gift recipients

**How it works:**
- For each purchase, the system randomly selects one type from the `type` array
- Depending on the type, a random ID is selected from the corresponding array (`user_receiver_id` or `channel_receiver_id`)
- This allows distributing gifts among multiple recipients of different types
- If you specify `type: [0]`, all gifts will be sent to yourself (IDs are ignored)

#### ⏱️ Check Frequency (`ticker`)

```json
{
    "ticker": 2.0
}
```

Interval between gift checks in seconds (default 2.0 seconds)

#### 🚀 Performance Parameters

New parameters for system performance optimization:

```json
{
    "retry_count": 4,
    "retry_delay": 0.01,
    "concurrency_gift_count": 10,
    "concurrent_operations": 300,
    "rpc_rate_limit": 30
}
```

- **`retry_count`** — number of retry attempts for failed purchases (default 4)
- **`retry_delay`** — delay between retry attempts in seconds (default 0.01 = 10ms)
- **`concurrency_gift_count`** — maximum number of gifts processed simultaneously (default 10)
- **`concurrent_operations`** — maximum number of concurrent operations (default 300)
- **`rpc_rate_limit`** — RPC request rate limit per second for Telegram API (default 30 RPS)

**Performance Optimization:**
- System now uses asynchronous retry processing
- RPC requests are limited to 30 per second to comply with Telegram limits
- Parallel processing of up to 300 operations simultaneously
- Minimal delay between retries (10ms) for maximum speed

#### 🧪 Test Mode (`test_mode`)

```json
{
    "test_mode": true
}
```

In test mode:
- **Total supply is ignored** (`total_supply`)
- **Capitalization is ignored**
- **Limited field should be negative** to buy unlimited gifts
- Gifts are purchased without real restrictions

#### 🔒 Global Limiters

```json
{
    "total_star_cap": 10000,
    "max_buy_count": 5,
    "limited_status": false
}
```

- **`total_star_cap`** — maximum stars to spend
- **`max_buy_count`** — global purchase limit
- **`limited_status`** — filter by gift limited status (true = only limited gifts, false = all gifts)

**Global limiter example:**
If you have criteria for 3+2+1=6 gifts, but `max_buy_count: 4`, only 4 gifts will be purchased (in order of appearance).

### 📋 Complete Configuration Example

```json
{
    "logger_level": "info",
    "soft_config": {
        "tg_settings": {
            "app_id": 12345678,
            "api_hash": "your_api_hash",
            "phone": "+1234567890",
            "password": "",
            "tg_bot_key": "",
            "notification_chat_id": 123456789
        },
        "criterias": [
            {
                "min_price": 10,
                "max_price": 50,
                "total_supply": 1000,
                "count": 3
            },
            {
                "min_price": 100,
                "max_price": 500,
                "total_supply": 500,
                "count": 2
            }
        ],
        "total_star_cap": 5000,
        "receiver": {
            "type": [0, 1],
            "user_receiver_id": [0, 123456789, 987654321],
            "channel_receiver_id": []
        },
        "test_mode": false,
        "max_buy_count": 4,
        "ticker": 2.0,
        "retry_count": 4,
        "retry_delay": 0.01,
        "limited_status": false,
        "concurrency_gift_count": 10,
        "concurrent_operations": 300,
        "rpc_rate_limit": 30
    }
}
```

### 📋 Requirements

- Go 1.23.4+
- Telegram account with API credentials
- Telegram bot for notifications (optional)

---

## ⚠️ Disclaimer

This software is provided "as is" for educational purposes. Users are responsible for compliance with Telegram's Terms of Service and any financial transactions performed by the software.

## 📄 License

This project is provided as-is for educational and personal use.
