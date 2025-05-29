package main

import (
	"context"
	"log"
	"os"

	"may_bot/deepseek"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Получаем токен бота из переменной окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	// Инициализируем бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	// Инициализируем DeepSeek клиент
	deepseekClient, err := deepseek.NewClient()
	if err != nil {
		log.Printf("Warning: DeepSeek client initialization failed: %v", err)
	}

	// Устанавливаем режим отладки
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Настраиваем получение обновлений
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	// Обрабатываем входящие сообщения
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Создаем ответное сообщение
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Обрабатываем команды
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg.Text = "Привет! Я бот с интеграцией DeepSeek. Задайте мне любой вопрос, и я постараюсь на него ответить. Используйте /help для получения списка команд."
			case "help":
				msg.Text = "Доступные команды:\n/start - Начать работу с ботом\n/help - Показать это сообщение\n/reset - Очистить историю нашего диалога\n\nПросто напишите свой вопрос, и я отвечу на него с помощью DeepSeek."
			case "reset":
				deepseekClient.ResetChatHistory(update.Message.Chat.ID)
				msg.Text = "История нашего диалога очищена. Можем начать новый разговор!"
			default:
				msg.Text = "Неизвестная команда. Используйте /help для получения списка команд."
			}
		} else {
			// Показываем, что бот печатает
			chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
			bot.Send(chatAction)

			// Создаем сообщение "бот думает"
			thinkingMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "🤔 Думаю...")
			sentMsg, err := bot.Send(thinkingMsg)
			if err != nil {
				log.Printf("Error sending thinking message: %v", err)
				continue
			}

			// Обрабатываем обычные сообщения через DeepSeek
			response, err := deepseekClient.GetResponse(context.Background(), update.Message.Chat.ID, update.Message.Text)
			if err != nil {
				// Удаляем сообщение "бот думает"
				deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, sentMsg.MessageID)
				bot.Send(deleteMsg)

				msg.Text = "Извините, произошла ошибка при обработке вашего запроса."
				log.Printf("Error getting DeepSeek response: %v", err)
			} else {
				// Удаляем сообщение "бот думает"
				deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, sentMsg.MessageID)
				bot.Send(deleteMsg)

				msg.Text = response
			}
		}

		// Отправляем сообщение
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
