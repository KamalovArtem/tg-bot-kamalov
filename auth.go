package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AuthStatus — статус авторизации пользователя
type AuthStatus struct {
	Status       string `json:"status"`        // success / expired / pending
	AccessToken  string `json:"access_token"`  // access-токен
	RefreshToken string `json:"refresh_token"` // refresh-токен
}

// RefreshResp — ответ на обновление токена
type RefreshResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// StartOAuthResp — ответ запуска OAuth
type StartOAuthResp struct {
	URL string `json:"url"` // ссылка для авторизации
}

// StartCodeResp — ответ запуска входа по коду
type StartCodeResp struct {
	Code string `json:"code"` // код для входа
}

// StartOAuth — начать OAuth-авторизацию (GitHub / Yandex)
func StartOAuth(authURL, provider, loginToken string) (string, int) {
	u := fmt.Sprintf("%s/auth/oauth/start?provider=%s&token_login=%s", authURL, provider, loginToken)
	resp, err := http.Post(u, "application/json", nil)
	if err != nil {
		return "", 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", resp.StatusCode
	}

	var data StartOAuthResp
	_ = json.NewDecoder(resp.Body).Decode(&data)
	return data.URL, 200
}

// StartCode — начать авторизацию по коду
func StartCode(authURL, loginToken string) (string, int) {
	u := fmt.Sprintf("%s/auth/code/start?token_login=%s", authURL, loginToken)
	resp, err := http.Post(u, "application/json", nil)
	if err != nil {
		return "", 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", resp.StatusCode
	}

	var data StartCodeResp
	_ = json.NewDecoder(resp.Body).Decode(&data)
	return data.Code, 200
}

// CheckAuthStatus — проверить статус авторизации
func CheckAuthStatus(authURL, loginToken string) (*AuthStatus, error) {
	resp, err := http.Get(authURL + "/auth/status?token_login=" + loginToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data AuthStatus
	err = json.NewDecoder(resp.Body).Decode(&data)
	return &data, err
}

// RefreshToken — обновить access-токен по refresh-токену
func RefreshToken(authURL, refreshToken string) (*RefreshResp, int) {
	body := fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken)

	resp, err := http.Post(
		authURL+"/auth/refresh",
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		return nil, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, resp.StatusCode
	}

	var data RefreshResp
	json.NewDecoder(resp.Body).Decode(&data)
	return &data, 200
}

// LogoutAll — выйти из системы на всех устройствах
func LogoutAll(authURL, refreshToken string) {
	body := fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken)
	http.Post(authURL+"/auth/logout", "application/json", strings.NewReader(body))
}
