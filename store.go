package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Контекст для работы с Redis
var ctx = context.Background()

// Возможные состояния пользователя
const (
	StateAnonymous  = "anonymous"  // пользователь начал вход
	StateAuthorized = "authorized" // пользователь авторизован
)

// Store — обёртка над Redis
type Store struct {
	rdb *redis.Client
}

// NewStore — создаёт подключение к Redis
func NewStore(addr string) *Store {
	return &Store{
		rdb: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

// key — формирует ключ Redis по chatID
func key(chatID int64) string {
	return fmt.Sprintf("tg:%d", chatID)
}

// GetAll — получить все данные пользователя по chatID
func (s *Store) GetAll(chatID int64) map[string]string {
	data, _ := s.rdb.HGetAll(ctx, key(chatID)).Result()
	return data
}

// Set — сохранить данные пользователя по chatID
func (s *Store) Set(chatID int64, data map[string]string) {
	s.rdb.HSet(ctx, key(chatID), data)
}

// Del — удалить данные пользователя по chatID
func (s *Store) Del(chatID int64) {
	s.rdb.Del(ctx, key(chatID))
}

// Keys — получить все ключи пользователей (tg:*)
func (s *Store) Keys() []string {
	keys, _ := s.rdb.Keys(ctx, "tg:*").Result()
	return keys
}

// GetByKey — получить одно поле по Redis-ключу
func (s *Store) GetByKey(k, field string) string {
	val, _ := s.rdb.HGet(ctx, k, field).Result()
	return val
}

// SetByKey — установить несколько полей по Redis-ключу
func (s *Store) SetByKey(k string, data map[string]string) {
	s.rdb.HSet(ctx, k, data)
}

// DelByKey — удалить данные пользователя по Redis-ключу
func (s *Store) DelByKey(k string) {
	s.rdb.Del(ctx, k)
}
