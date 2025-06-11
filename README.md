# Gift Buyer

[![Language: Russian](https://img.shields.io/badge/Language-Русский-blue)](#русский) [![Language: English](https://img.shields.io/badge/Language-English-green)](#english)

---

## Русский

### 📖 Описание

**Gift Buyer** — автоматизированная система для покупки Star Gifts в Telegram. Программа непрерывно мониторит доступные подарки, проверяет их соответствие заданным критериям и автоматически покупает подходящие варианты.

### ⚡ Преимущества

- **🚀 Высокая скорость** — мгновенная реакция на появление новых подарков
- **🎯 Точная фильтрация** — настраиваемые критерии по цене, количеству и лимитам
- **⚙️ Параллельная обработка** — одновременная покупка нескольких подарков
- **🔄 Автоматические повторы** — система retry при ошибках
- **📱 Уведомления** — мгновенные оповещения в Telegram
- **💾 Кэширование** — сохранение состояния между перезапусками
- **🛡️ Безопасность** — graceful shutdown и обработка ошибок

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
        "receiver_id": [123456789, 987654321, 555666777]
    }
}
```

- **`type`** — массив типов получателей:
  - `0` — отправить себе
  - `1` — отправить другому пользователю (пользователь должен быть в ваших контактах)
  - `2` — отправить в канал (канал должен принадлежать вашему аккаунту)
- **`receiver_id`** — массив ID получателей (пользователей или каналов)

**Как это работает:**
- При каждой покупке система случайно выбирает один тип из массива `type` и один ID из массива `receiver_id`
- Это позволяет распределять подарки между несколькими получателями
- Если указать `type: [0]` и `receiver_id: [0]`, все подарки будут отправляться себе

#### ⏱️ Частота проверки (`check_interval_ms`)

```json
{
    "check_interval_ms": 1000
}
```

Интервал между проверками новых подарков в миллисекундах (по умолчанию 1000мс = 1 секунда)

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
    "max_buy_count": 5
}
```

- **`total_star_cap`** — максимальная капитализация подарка в звездах
- **`max_buy_count`** — глобальный лимит покупок

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
            "receiver_id": [0, 123456789, 987654321]
        },
        "test_mode": false,
        "max_buy_count": 4,
        "check_interval_ms": 500
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
- **⚙️ Parallel Processing** — simultaneous purchase of multiple gifts
- **🔄 Auto Retry** — retry system for error handling
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
        "receiver_id": [123456789, 987654321, 555666777]
    }
}
```

- **`type`** — array of recipient types:
  - `0` — send to yourself
  - `1` — send to another user (user must be in your contacts)
  - `2` — send to channel (channel must belong to your account)
- **`receiver_id`** — array of recipient IDs (users or channels)

**How it works:**
- For each purchase, the system randomly selects one type from the `type` array and one ID from the `receiver_id` array
- This allows distributing gifts among multiple recipients
- If you specify `type: [0]` and `receiver_id: [0]`, all gifts will be sent to yourself

#### ⏱️ Check Frequency (`check_interval_ms`)

```json
{
    "check_interval_ms": 1000
}
```

Interval between gift checks in milliseconds (default 1000ms = 1 second)

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
    "max_buy_count": 5
}
```

- **`total_star_cap`** — maximum stars to spend
- **`max_buy_count`** — global purchase limit

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
            "receiver_id": [0, 123456789, 987654321]
        },
        "test_mode": false,
        "max_buy_count": 4,
        "check_interval_ms": 500
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
