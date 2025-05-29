# Telegram Bot

Telegram бот с интеграцией DeepSeek AI, написанный на Go.

## Требования

- Go 1.16 или выше
- Telegram Bot Token (получить у [@BotFather](https://t.me/BotFather))
- DeepSeek API Key

## Установка

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd may_bot
```

2. Установите зависимости:
```bash
go mod download
```

## Запуск

1. Установите переменные окружения:
```bash
export TELEGRAM_BOT_TOKEN="ваш_токен_бота"
export DEEPSEEK_API_KEY="ваш_ключ_api_deepseek"
```

2. Запустите бота:
```bash
go run main.go
```

## Доступные команды

- `/start` - Начать работу с ботом
- `/help` - Показать список доступных команд

## Функциональность

- Бот отвечает на команды /start и /help
- Все текстовые сообщения обрабатываются через DeepSeek AI
- Бот предоставляет интеллектуальные ответы на вопросы пользователей