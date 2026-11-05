Вариант 2 — это переход к использованию **автоинкрементного числового поля** в самой базе данных. Это более «взрослый» подход, но он потребует изменений в схеме и логике.

### 🧠 Суть Варианта 2

Вместо того чтобы хранить `uuid` как текст (который может быть `"A"`, `"1"` и т.д.), вы добавляете в таблицу `urls` колонку `id SERIAL PRIMARY KEY`. База данных сама будет генерировать уникальные числа (1, 2, 3...). А ваш `shortener` будет использовать это число для генерации Base62-кода.

### 🛠️ Шаг 1. Обновите схему таблицы

В `NewPostgresStorage` измените `CREATE TABLE`:

```go
query := `
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    user_id TEXT
);`
```

Убрали `uuid TEXT PRIMARY KEY`, добавили `id SERIAL`.

### 🛠️ Шаг 2. Обновите метод `Save` в `PostgresStorage`

Теперь `Save` должен **не передавать `uuid`**, а позволить базе сгенерировать `id`. Но тогда как получить `short_url`? Вам нужно сначала вставить запись с `original_url` и `user_id`, получить `id`, а затем сгенерировать `short_url` на основе `id` и обновить запись.

**Вариант А (два запроса):**

```go
func (s *PostgresStorage) Save(code, longURL, userID string) error {
    // code уже сгенерирован shortener'ом, но мы его не используем для id.
    // Мы хотим, чтобы id генерировался БД, а short_url = FromBase62(id).
    // Это не совсем то, что нужно. 
}
```

Этот путь сложен, потому что `shortener.Shorten()` генерирует код **до** вставки. Если вы хотите, чтобы `id` генерировался БД, вам нужно изменить логику: сначала вставить, потом обновить.

**Вариант Б (изменить `shortener`, чтобы он не генерировал код, а брал `id` из БД):**

Это глобальное изменение. Возможно, для курса это избыточно.

### 🛠️ Шаг 3. Альтернатива: добавить числовое поле, но оставить генерацию в Go

Вы можете добавить в таблицу поле `counter BIGINT` (или `num_id`), которое будет хранить числовое значение Base62-кода. Тогда:

1.  В `Save` вы передаете `counter` (число, которое вернул `shortener`).
2.  `GetMaxUUID` делает `SELECT MAX(counter) FROM urls`.
3.  В `main.go` вызываете `shortener.SetCounter(maxCounter)`.

Это самый простой и совместимый с текущим кодом вариант.

### 📄 Как это выглядит

**Схема:**
```sql
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    counter BIGINT NOT NULL UNIQUE, -- числовое значение кода
    short_url TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    user_id TEXT
);
```

**`Save`:**
```go
func (s *PostgresStorage) Save(code string, counter int64, longURL, userID string) error {
    query := `INSERT INTO urls (counter, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)`
    _, err := s.pool.Exec(context.Background(), query, counter, code, longURL, userID)
    return err
}
```

**`GetMaxCounter`:**
```go
func (s *PostgresStorage) GetMaxCounter(ctx context.Context) (int64, error) {
    var maxCounter int64
    query := `SELECT COALESCE(MAX(counter), 0) FROM urls`
    err := s.pool.QueryRow(ctx, query).Scan(&maxCounter)
    return maxCounter, err
}
```

**В `main.go`:**
```go
maxCounter, err := pgRepo.GetMaxCounter(context.Background())
if err != nil { ... }
shortener.SetCounter(int(maxCounter))
```

### 💎 Итог

> **Вариант 2 (с `SERIAL`) требует переписывания логики генерации кода, что сложно и не нужно для курса.** Проще добавить в таблицу числовое поле `counter` (или `num_id`), которое хранит число из Base62. Тогда `MAX(counter)` работает без ошибок, а `shortener` продолжает генерировать коды как раньше. Это гибридный вариант, который решает проблему с `"A"` и сохраняет обратную совместимость.