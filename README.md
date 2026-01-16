# Telegram Module — TestFlow Assistant

Telegram-модуль проекта **TestFlow**. Реализован как простой клиент между Telegram и backend-сервисами системы онлайн‑тестирования.

Модуль не содержит бизнес‑логики — он отвечает только за:

* взаимодействие с Telegram API
* авторизацию пользователей
* проксирование команд в `main-module`
* получение уведомлений

---

## 🚀 Возможности

* 🔐 Авторизация через **GitHub**, **Yandex** или **код**
* 📚 Работа с пользователями, дисциплинами, тестами
* 🧪 Запуск и завершение попыток прохождения тестов
* 🔔 Автоматические уведомления
* ♻️ Авто‑обновление access‑токена по refresh‑токену

Все команды из `main-module` поддерживаются через проксирование.

---

## 💬 Основные команды
```
/start
/help
```

### Авторизация

```
/login
/login github
/login yandex
/login code
/logout
/logout all=true
```

### Команды для main‑module

```
/users_list
/user_get <id>
/user_fullname_set <id> "<fullName>"
/user_data <id>
/user_roles_get <id>
/user_roles_set <id> <rolesCsv>
/user_block_get <id>
/user_block_set <id> <true|false>

/courses_list
/course_get <course_id>
/course_create "<name>" "<description>" <teacher_id>
/course_update <course_id> "<name>" "<description>"
/course_delete <course_id>
/course_students <course_id>
/course_student_add <course_id> <user_id>
/course_student_del <course_id> <user_id>

/questions_list
/question_get <id>
/question_create "<title>" "<text>" "<optionsCsv>" <correctIndex>
/question_update <id> "<title>" "<text>" "<optionsCsv>" <correctIndex>
/question_delete <id>

/test_attempts <test_id>
/test_user_grade <test_id> <user_id>
/test_user_answers <test_id> <user_id>

/attempt_create <test_id>
/attempt_get <attempt_id>
/attempt_finish <attempt_id>

/answers_get <attempt_id>
/answer_update <answer_id> <index>
/answer_delete <answer_id>

/notification
/notification_delete
```



---
## 🔐 Переменные окружения

В проекте используется файл `.env` с реальными значениями  
(токен Telegram-бота, адреса сервисов и т.д.).

По соображениям безопасности данный файл **не публикуется в репозитории**,  
чтобы исключить утечку чувствительных данных и предотвратить  
несанкционированное использование бота.

Вместо него в репозитории представлен файл `.env.example`,  
который демонстрирует структуру и список необходимых  
переменных окружения для запуска проекта.

---

## 🧠 Архитектура

Telegram‑модуль:

* хранит состояние пользователя в **Redis**
* получает `access_token` / `refresh_token` от Auth‑модуля
* отправляет HTTP‑запросы в **main‑module**
* возвращает пользователю ответ сервера

Все проверки прав и бизнес‑логика находятся в `main-module`.

---

## ⚙️ Запуск

### Запуск через Docker

```
docker compose up --build
```

---

## 📌 Примечания

* Telegram‑бот принимает команды в текстовом виде (без inline‑кнопок)
* Ошибки `403 / 422 / 500` возвращаются напрямую от `main-module`
* Модуль реализован максимально просто и логично

---

## 👨‍💻 Автор

Telegram‑модуль разрабатывал: Камалов Артём Ильгизович

**Команда:** Enigma Team
