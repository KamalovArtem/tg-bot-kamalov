package main

import (
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Таймер №1: проверка входа (Anonymous)
func startLoginTimer(bot *tgbotapi.BotAPI, store *Store, authURL string) {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for range ticker.C {
			keys := store.Keys()

			for _, k := range keys {
				state := store.GetByKey(k, "state")
				if state != StateAnonymous {
					continue
				}

				loginToken := store.GetByKey(k, "login_token")
				if loginToken == "" {
					continue
				}

				st, err := CheckAuthStatus(authURL, loginToken)
				if err != nil {
					continue
				}

				chatID := keyToChatID(k)

				switch st.Status {
				case "success":
					// сохраняем токены, переводим в Authorized
					store.SetByKey(k, map[string]string{
						"state":         StateAuthorized,
						"access_token":  st.AccessToken,
						"refresh_token": st.RefreshToken,
					})

					bot.Send(tgbotapi.NewMessage(chatID, "✅ Авторизация успешна!"))

				case "expired":
					// удаляем сессию
					store.DelByKey(k)
					bot.Send(tgbotapi.NewMessage(chatID, "❌ Авторизация не удалась или истекла. Используйте /login"))
				}
			}
		}
	}()
}

// Таймер №2: уведомления (Authorized) + refresh на 401
func startNotificationTimer(bot *tgbotapi.BotAPI, store *Store, authURL, mainURL string) {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for range ticker.C {
			keys := store.Keys()

			for _, k := range keys {
				state := store.GetByKey(k, "state")
				if state != StateAuthorized {
					continue
				}

				accessToken := store.GetByKey(k, "access_token")
				refreshToken := store.GetByKey(k, "refresh_token")
				if accessToken == "" || refreshToken == "" {
					continue
				}

				chatID := keyToChatID(k)

				// 1) пробуем получить уведомления
				notifs, code := GetNotifications(mainURL, accessToken)

				// 2) если access умер — делаем refresh и повторяем 1 раз
				if code == 401 {
					newTokens, rc := RefreshToken(authURL, refreshToken)
					if rc != 200 {
						// refresh умер -> разлогиниваем
						store.DelByKey(k)
						bot.Send(tgbotapi.NewMessage(chatID, "Вы не авторизованы. Используйте /login"))
						continue
					}

					// сохраняем новые токены
					store.SetByKey(k, map[string]string{
						"access_token":  newTokens.AccessToken,
						"refresh_token": newTokens.RefreshToken,
					})

					// повторяем запрос уведомлений с новым access
					notifs, code = GetNotifications(mainURL, newTokens.AccessToken)
				}

				// 3) нет уведомлений или ошибка
				if code != 200 || len(notifs) == 0 {
					continue
				}

				// 4) отправляем уведомления
				for _, n := range notifs {
					bot.Send(tgbotapi.NewMessage(chatID, "🔔 "+n.Payload))
				}

				// 5) удаляем уведомления в main-module
				if code == 200 {
					// используй актуальный access
					DeleteNotifications(mainURL, store.GetByKey(k, "access_token"))
				}

			}
		}
	}()
}

// keyToChatID — извлекает chatID из Redis-ключа вида "tg:<id>"
func keyToChatID(k string) int64 {
	s := strings.TrimPrefix(k, "tg:")
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}
