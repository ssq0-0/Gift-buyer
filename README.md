# Gift Buyer - Система покупки Telegram подарков

[![Language: Russian](https://img.shields.io/badge/Language-Русский-blue)](#русский) [![Language: English](https://img.shields.io/badge/Language-English-green)](#english)

---

## Русский

### 📖 Описание

**Gift Buyer** — автоматизированная система для покупки Star Gifts в Telegram. Программа непрерывно мониторит доступные подарки, проверяет их соответствие заданным критериям и автоматически покупает подходящие варианты.

### ⚡ Преимущества

- **🚀 Высокая скорость** — мгновенная реакция на появление новых подарков
- **🎯 Точная фильтрация** — настраиваемые критерии по цене, количеству и лимитам
- **⚙️ Параллельная обработка** — одновременная покупка нескольких подарков (до 300 операций)
- **🔄 Асинхронные повторы** — интеллектуальная система retry с минимальными задержками (10мс)
- **⚡ Оптимизированная производительность** — RPC лимит 30 запросов/сек для максимальной скорости
- **📱 Уведомления** — мгновенные оповещения в Telegram
- **💾 Кэширование** — сохранение состояния между перезапусками
- **🛡️ Безопасность** — graceful shutdown и обработка ошибок

### 🚀 Быстрый старт

### Установка зависимостей
```bash
make deps
```

### Настройка переменных окружения
Скопируйте файл с примером переменных окружения:
```bash
cp env.example .env
```

Заполните `.env` файл своими данными:
- `TG_APP_ID` - ID приложения Telegram
- `TG_API_HASH` - API Hash от Telegram
- `TG_PHONE` - Номер телефона
- `TG_PASSWORD` - Пароль (опционально)

### Сборка и запуск
```bash
make build
make run
```

## 🛠 Разработка

### Доступные команды
```bash
make help          # Показать все доступные команды
make all           # Полная проверка проекта (рекомендуется перед коммитом)
make test          # Запустить тесты
make lint          # Проверить код линтером
make fmt           # Отформатировать код
make build         # Собрать бинарный файл
make security      # Проверить безопасность кода
make coverage      # Запустить тесты с покрытием
```

### Установка инструментов разработки
```bash
make install-tools
```

## 🔧 CI/CD Pipeline

Проект использует GitHub Actions для автоматизации:

### Этапы pipeline:
1. **Test** - Тестирование и проверка качества кода
2. **Build** - Сборка для разных платформ
3. **Release** - Создание релизов (только для main ветки)
4. **Security** - Проверка безопасности

### Локальная проверка перед коммитом:
```bash
make all
```

Эта команда выполнит все проверки, которые будут запущены в CI/CD.

## 📁 Структура проекта

```
├── cmd/                    # Точки входа приложения
├── internal/              # Внутренняя логика
│   ├── config/           # Конфигурация
│   └── service/          # Бизнес-логика
├── pkg/                  # Переиспользуемые пакеты
├── .github/workflows/    # GitHub Actions
├── docs/                 # Документация
├── Makefile             # Команды для разработки
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
make test           # Запустить все тесты
make coverage       # Тесты с отчетом о покрытии
```

Отчет о покрытии сохраняется в `coverage.html`.

## 🚀 Деплой

### Автоматический деплой
При пуше в `main` ветку автоматически:
1. Запускаются все тесты
2. Собираются бинарные файлы для всех платформ
3. Создается релиз с артефактами

### Ручная сборка
```bash
make build-all      # Сборка для всех платформ
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
3. Запустите `make all` для проверки
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
        "type": [1, 2],
        "user_receiver_id": [123456789, 987654321],
        "channel_receiver_id": [555666777, 888999000]
    }
}
```

- **`type`** — массив типов получателей:
  - `0` — отправить себе
  - `1` — отправить другому пользователю (пользователь должен быть в ваших контактах)
  - `2` — отправить в канал (канал должен принадлежать вашему аккаунту)
- **`user_receiver_id`** — массив ID пользователей для получения подарков
- **`channel_receiver_id`** — массив ID каналов для получения подарков

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
    "retry_count": 4,
    "retry_delay": 0.01,
    "concurrency_gift_count": 10,
    "concurrent_operations": 300,
    "rpc_rate_limit": 30
}
```

- **`retry_count`** — количество попыток повтора при неудачной покупке (по умолчанию 4)
- **`retry_delay`** — задержка между попытками повтора в секундах (по умолчанию 0.01 = 10мс)
- **`concurrency_gift_count`** — максимальное количество подарков, обрабатываемых одновременно (по умолчанию 10)
- **`concurrent_operations`** — максимальное количество одновременных операций (по умолчанию 300)
- **`rpc_rate_limit`** — лимит RPC запросов в секунду для Telegram API (по умолчанию 30 RPS)

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
- Telegram аккаунт с API ключами
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

1. **Download and build the project:**
   ```bash
   git clone <repository-url>
   cd gift-buyer
   go build -o gift-buyer cmd/main.go
   ```

2. **Configure the application:**
   ```bash
   cp internal/config/config_example.json internal/config/config.json
   # Edit config.json with your credentials
   ```

3. **Run the program:**
   ```bash
   ./gift-buyer
   ```

### ⚙️ Detailed Configuration

#### 🔧 Telegram Settings (`tg_settings`)

```json
{
    "tg_settings": {
        "app_id": 12345678,
        "api_hash": "your_api_hash",
        "phone": "+1234567890",
        "password": "2fa_password",
        "tg_bot_key": "bot_token",
        "notification_chat_id": 123456789
    }
}
```

- **`app_id`** and **`api_hash`** — required parameters from [my.telegram.org](https://my.telegram.org)
- **`phone`** — account phone number in international format
- **`password`** — two-factor authentication password (can be left empty `""` if 2FA is disabled)
- **`tg_bot_key`** — Telegram bot token for notifications (can be left empty `""` if notifications not needed)
- **`notification_chat_id`** — chat ID for sending notifications

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
