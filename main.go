package main

import (
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Создаём Telegram-бота по токену из переменных окружения
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_TOKEN"))
	if err != nil {
		panic(err)
	}

	// Инициализируем хранилище состояний пользователей (Redis)
	store := NewStore(os.Getenv("REDIS_ADDR"))

	// Адреса внешних сервисов
	authURL := os.Getenv("AUTH_URL")
	mainURL := os.Getenv("MAIN_URL")

	// Запускаем таймер проверки авторизации
	startLoginTimer(bot, store, authURL)

	// Запускаем таймер уведомлений и refresh токена
	startNotificationTimer(bot, store, authURL, mainURL)

	// Настраиваем получение обновлений от Telegram
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Основной цикл обработки сообщений
	for update := range bot.GetUpdatesChan(u) {
		if update.Message == nil {
			continue
		}
		// Обрабатываем входящее сообщение пользователя
		HandleMessage(bot, store, authURL, update.Message)
	}
}
