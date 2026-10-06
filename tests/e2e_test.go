package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// Константа с адресом запущенного сервера (в Docker)
const apiURL = "http://localhost:8080/api"

// Глобальные переменные для сохранения состояния между тестами
var (
	authToken string
	postID    int
)

// Вспомогательная функция для генерации уникального имени (чтобы тесты не падали при перезапуске)
func getUniqueUsername() string {
	return fmt.Sprintf("test_user_%d", time.Now().UnixNano())
}

// 1. Проверка доступности API
func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(apiURL + "/health")
	if err != nil {
		t.Fatalf("Сервер недоступен (не забудьте запустить docker-compose up -d): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус 200, получено: %d", resp.StatusCode)
	}
}

// 2. Регистрация и Логин (Авторизация)
func TestAuth(t *testing.T) {
	username := getUniqueUsername()
	password := "supersecret123"

	reqBody := map[string]string{
		"username": username,
		"password": password,
	}
	bodyData, _ := json.Marshal(reqBody)

	// Регистрация
	respReg, err := http.Post(apiURL+"/register", "application/json", bytes.NewBuffer(bodyData))
	if err != nil {
		t.Fatalf("Ошибка при регистрации: %v", err)
	}
	defer respReg.Body.Close()
	if respReg.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200 при регистрации, получено: %d", respReg.StatusCode)
	}

	// Логин
	respLog, err := http.Post(apiURL+"/login", "application/json", bytes.NewBuffer(bodyData))
	if err != nil {
		t.Fatalf("Ошибка при логине: %v", err)
	}
	defer respLog.Body.Close()

	if respLog.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200 при логине, получено: %d", respLog.StatusCode)
	}

	var respData map[string]string
	json.NewDecoder(respLog.Body).Decode(&respData)

	token, ok := respData["token"]
	if !ok || token == "" {
		t.Fatal("Токен не получен")
	}
	
	authToken = token // сохраняем для следующих тестов
	t.Logf("Успешная регистрация и авторизация. Токен: %s...", token[:10])
}

// 3. Получение списка категорий
func TestCategories(t *testing.T) {
	resp, err := http.Get(apiURL + "/categories")
	if err != nil {
		t.Fatalf("Ошибка при получении категорий: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получено: %d", resp.StatusCode)
	}

	var categories []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&categories)

	if len(categories) == 0 {
		t.Error("Список категорий пуст, хотя скрипт init.sql должен был их создать")
	}
	t.Logf("Получено %d категорий", len(categories))
}

// 4. Создание нового поста
func TestCreatePost(t *testing.T) {
	if authToken == "" {
		t.Skip("Пропуск теста, так как нет токена (TestAuth упал)")
	}

	reqBody := map[string]interface{}{
		"category_id": 2, // Категория "Игры"
		"title":       "Интеграционный тест Go",
		"content":     "Этот пост создан автоматически при тестировании",
	}
	bodyData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", apiURL+"/posts", bytes.NewBuffer(bodyData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Ошибка при создании поста: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Ожидался статус 201, получено: %d. Тело: %s", resp.StatusCode, string(body))
	}

	var respData map[string]int
	json.NewDecoder(resp.Body).Decode(&respData)

	id, ok := respData["id"]
	if !ok || id == 0 {
		t.Fatal("Не получен ID поста")
	}
	postID = id
	t.Logf("Пост успешно создан, ID = %d", postID)
}

// 5. Создание комментария к посту
func TestCreateComment(t *testing.T) {
	if authToken == "" || postID == 0 {
		t.Skip("Пропуск теста (нет токена или postID)")
	}

	reqBody := map[string]interface{}{
		"content": "Согласен, всё работает отлично!",
	}
	bodyData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/posts/%d/comments", apiURL, postID)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Ошибка при создании комментария: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Ожидался статус 201, получено: %d", resp.StatusCode)
	}
	t.Log("Комментарий успешно создан")
}

// 6. Получение ленты постов (Hot Feed)
func TestHotFeed(t *testing.T) {
	resp, err := http.Get(apiURL + "/posts")
	if err != nil {
		t.Fatalf("Ошибка получения ленты: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получено: %d", resp.StatusCode)
	}

	var posts []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&posts)

	if len(posts) == 0 {
		t.Error("Лента пуста, но мы только что создали пост")
	} else {
		t.Logf("Получено %d постов в ленте", len(posts))
	}
}
