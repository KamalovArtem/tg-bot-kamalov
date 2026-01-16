package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleMessage — главный обработчик сообщений Telegram
// Управляет логикой в зависимости от состояния пользователя
func HandleMessage(bot *tgbotapi.BotAPI, store *Store, authURL string, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)

	// Команда и первый аргумент
	cmd, arg := parseCmd(text)

	// Достаём данные пользователя из Redis
	data := store.GetAll(chatID)
	state := data["state"]

	// --- Пользователь ещё не начинал авторизацию
	if state == "" {
		switch cmd {
		case "login":
			// /login без параметров
			if arg == "" {
				send(bot, chatID, "Вы не авторизованы.\n/login github\n/login yandex\n/login code")
				return
			}
			// /login <type>
			startLogin(bot, store, authURL, chatID, arg)
			return

		case "logout":
			send(bot, chatID, "Вы не авторизованы.\n/login github\n/login yandex\n/login code")
			return

		case "help", "start":
			send(bot, chatID, "Команды:\n/login\n/login github|yandex|code\n/logout\n/help")
			return

		default:
			send(bot, chatID, "Нет такой команды")
			return
		}
	}

	// --- Пользователь в процессе авторизации
	if state == StateAnonymous {
		switch cmd {
		case "login":
			// /login <type> — перегенерить login_token и заново стартануть
			if arg != "" {
				startLogin(bot, store, authURL, chatID, arg)
				return
			}
			// /login без параметров — как “любой другой”: проверяем статус
			checkAnonymousStatus(bot, store, authURL, chatID)
			return

		case "logout":
			store.Del(chatID)
			send(bot, chatID, "Сеанс завершён")
			return

		case "help", "start":
			send(bot, chatID, "Вы в процессе входа.\n/login github|yandex|code\n/logout\n/help")
			return

		default:
			// “любой другой” → проверить /auth/status по login_token
			if cmd == "" {
				send(bot, chatID, "Нет такой команды")
				return
			}
			checkAnonymousStatus(bot, store, authURL, chatID)
			return
		}
	}

	// --- Пользователь авторизован
	if state == StateAuthorized {
		switch cmd {
		case "login":
			send(bot, chatID, "Вы уже авторизованы ✅")
			return

		case "logout":
			// /logout all=true
			if strings.Contains(strings.ToLower(text), "all=true") {
				LogoutAll(authURL, data["refresh_token"])
				send(bot, chatID, "Сеанс завершён на всех устройствах")
			} else {
				send(bot, chatID, "Сеанс завершён")
			}
			store.Del(chatID)
			return

		case "help", "start":
			send(bot, chatID, "Вы авторизованы ✅\n/logout\n/help")
			return

		default:
			access := data["access_token"]

			// имя команды без /
			cmdName, _ := parseCmd(text)
			args := parseArgs(text)

			code, body := CallMain(getMainURL(), access, cmdName, args)
			if code == 0 {
				send(bot, chatID, "Нет такой команды")
				return
			}

			send(bot, chatID, body)
			return

		}
	}

	// fallback
	send(bot, chatID, "Используйте /login")
}

// parseCmd — получить команду и первый аргумент из сообщения
func parseCmd(text string) (cmd string, arg string) {
	t := strings.TrimSpace(text)
	if t == "" {
		return "", ""
	}
	if !strings.HasPrefix(t, "/") {
		return "", ""
	}
	t = strings.TrimPrefix(t, "/")
	parts := strings.Fields(t)
	if len(parts) == 0 {
		return "", ""
	}
	cmd = strings.ToLower(parts[0])
	if len(parts) > 1 {
		arg = strings.ToLower(parts[1])
	}
	return cmd, arg
}

// startLogin — запуск авторизации (OAuth или код)
func startLogin(bot *tgbotapi.BotAPI, store *Store, authURL string, chatID int64, loginType string) {
	if loginType != "github" && loginType != "yandex" && loginType != "code" {
		send(bot, chatID, "Доступно: /login github, /login yandex, /login code")
		return
	}

	loginToken := fmt.Sprintf("%d", time.Now().UnixNano())

	store.Set(chatID, map[string]string{
		"state":       StateAnonymous,
		"login_token": loginToken,
	})

	if loginType == "code" {
		code, st := StartCode(authURL, loginToken)
		if st != 200 || code == "" {
			store.Del(chatID)
			send(bot, chatID, "Ошибка запуска входа через код")
			return
		}
		send(bot, chatID, "🔑 Код для входа: "+code+"\n⏳ Ожидаем подтверждение входа")
		return
	}

	url, st := StartOAuth(authURL, loginType, loginToken)
	if st != 200 || url == "" {
		store.Del(chatID)
		send(bot, chatID, "Ошибка запуска входа")
		return
	}

	send(bot, chatID, "Перейдите по ссылке:\n"+url+"\n⏳ Ожидаем подтверждение входа")
}

// checkAnonymousStatus — проверка статуса авторизации через Auth-модуль
func checkAnonymousStatus(bot *tgbotapi.BotAPI, store *Store, authURL string, chatID int64) {
	data := store.GetAll(chatID)
	loginToken := data["login_token"]
	if loginToken == "" {
		store.Del(chatID)
		send(bot, chatID, "Вы не авторизованы.\n/login github\n/login yandex\n/login code")
		return
	}

	st, err := CheckAuthStatus(authURL, loginToken)
	if err != nil {
		// Если Auth временно недоступен — просто ждём
		send(bot, chatID, "⏳ Ожидаем подтверждение входа")
		return
	}

	switch st.Status {
	case "success":
		//  Проверить что 2 токена есть
		if st.AccessToken == "" || st.RefreshToken == "" {
			store.Del(chatID)
			send(bot, chatID, "Ошибка авторизации: токены не получены")
			return
		}
		store.Set(chatID, map[string]string{
			"state":         StateAuthorized,
			"access_token":  st.AccessToken,
			"refresh_token": st.RefreshToken,
		})
		send(bot, chatID, "✅ Авторизация успешна!")

	case "expired":
		store.Del(chatID)
		send(bot, chatID, "Вы не авторизованы.\n/login github\n/login yandex\n/login code")

	case "denied":
		store.Del(chatID)
		send(bot, chatID, "❌ Неудачная авторизация")

	default:
		// pending / unknown — ждём
		send(bot, chatID, "⏳ Ожидаем подтверждение входа")
	}
}

// send — отправка сообщения в Telegram
func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	bot.Send(tgbotapi.NewMessage(chatID, text))
}

// getMainURL — адрес main-module
func getMainURL() string {
	return os.Getenv("MAIN_URL")
}

// parseArgs — парсит аргументы команды из текста сообщения
func parseArgs(text string) []string {
	t := strings.TrimSpace(strings.TrimPrefix(text, "/"))
	parts := splitArgs(t)
	if len(parts) <= 1 {
		return nil
	}
	return parts[1:]
}

// splitArgs — разбивает строку, поддерживает кавычки
func splitArgs(s string) []string {
	var res []string
	var cur strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuotes = !inQuotes
			continue
		}
		if c == ' ' && !inQuotes {
			if cur.Len() > 0 {
				res = append(res, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteByte(c)
	}
	if cur.Len() > 0 {
		res = append(res, cur.String())
	}
	return res
}
