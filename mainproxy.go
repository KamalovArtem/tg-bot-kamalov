package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// cmdSpec — описание команды main-module
// Method — HTTP метод, Params — список параметров
type cmdSpec struct {
	Method string
	Params []string
}

// // mainCmds — таблица всех поддерживаемых команд main-module
// Используется для проксирования команд из Telegram
var mainCmds = map[string]cmdSpec{
	"users_list":        {"GET", nil},
	"user_get":          {"GET", []string{"id"}},
	"user_fullname_set": {"POST", []string{"id", "fullName"}},
	"user_data":         {"GET", []string{"id"}},
	"user_roles_get":    {"GET", []string{"id"}},
	"user_roles_set":    {"POST", []string{"id", "rolesCsv"}},
	"user_block_get":    {"GET", []string{"id"}},
	"user_block_set":    {"POST", []string{"id", "blocked"}},

	"courses_list":       {"GET", nil},
	"course_get":         {"GET", []string{"course_id"}},
	"course_create":      {"POST", []string{"name", "description", "teacher_id"}},
	"course_update":      {"POST", []string{"course_id", "name", "description"}},
	"course_delete":      {"POST", []string{"course_id"}},
	"course_students":    {"GET", []string{"course_id"}},
	"course_student_add": {"POST", []string{"course_id", "user_id"}},
	"course_student_del": {"POST", []string{"course_id", "user_id"}},

	"questions_list":  {"GET", nil},
	"question_get":    {"GET", []string{"id"}},
	"question_create": {"POST", []string{"title", "text", "optionsCsv", "correctIndex"}},
	"question_update": {"POST", []string{"id", "title", "text", "optionsCsv", "correctIndex"}},
	"question_delete": {"POST", []string{"id"}},

	"attempt_create": {"POST", []string{"test_id"}},
	"attempt_get":    {"GET", []string{"attempt_id"}},
	"attempt_finish": {"POST", []string{"attempt_id"}},

	"answers_get":   {"GET", []string{"attempt_id"}},
	"answer_update": {"POST", []string{"answer_id", "index"}},
	"answer_delete": {"POST", []string{"answer_id"}},

	"notification":        {"GET", nil},
	"notification_delete": {"POST", nil},
}

// CallMain — универсальный прокси в main-module
// Формирует HTTP-запрос по имени команды и аргументам
func CallMain(mainURL, accessToken, cmd string, args []string) (int, string) {

	// Проверяем, поддерживается ли команда
	spec, ok := mainCmds[cmd]
	if !ok {
		return 0, ""
	}

	// Формируем query-параметры
	q := url.Values{}
	for i, p := range spec.Params {
		if i >= len(args) {
			return 422, "Not enough arguments"
		}
		q.Set(p, args[i])
	}

	// Собираем URL запроса
	u := strings.TrimRight(mainURL, "/") + "/" + cmd
	if len(q) > 0 {
		u += "?" + q.Encode()
	}

	// Создаём HTTP-запрос с access_token
	req, _ := http.NewRequest(spec.Method, u, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Отправляем запрос в main-module
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()

	// Возвращаем ответ сервера как текст
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}
