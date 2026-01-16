package main

import (
	"encoding/json"
	"net/http"
)

// Notification — структура уведомления из main-module
type Notification struct {
	ID      int    `json:"id"`
	Payload string `json:"payload"`
}

// GetNotifications — получить список уведомлений пользователя
// Делает запрос в main-module с access_token
func GetNotifications(mainURL, accessToken string) ([]Notification, int) {
	req, _ := http.NewRequest("GET", mainURL+"/notification", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0
	}
	defer resp.Body.Close()

	// Если сервер вернул ошибку — отдаём код
	if resp.StatusCode != 200 {
		return nil, resp.StatusCode
	}

	// Парсим JSON-ответ в список уведомлений
	var data []Notification
	json.NewDecoder(resp.Body).Decode(&data)
	return data, 200
}

// DeleteNotifications — удалить все уведомления пользователя
// Используется после успешной отправки уведомлений в Telegram
func DeleteNotifications(mainURL, accessToken string) {
	req, _ := http.NewRequest("POST", mainURL+"/notification_delete", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	http.DefaultClient.Do(req)
}
