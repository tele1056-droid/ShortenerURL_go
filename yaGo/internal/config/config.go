package config

import (
	"flag"
	"os"
)

// в этой структуре мы будем хранить переменные для аргументов команд. строки (флаги)
type Config struct {
	ServerAddress   string // адрес запуска HTTP-сервера (например, localhost:8080)
	BaseURL         string // базовый адрес результирующего URL (например, http://localhost:8080)
	FileStoragePath string // конфигурация для пути к файлу
}

// NewConfig читает флаги командной строки и возвращает готовую конфигурацию
func NewConfig() *Config {
	cfg := &Config{}

	// 1. Читаем флаги (с значениями по умолчанию)
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "базовый адрес результирующего URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "short-url-db.json", "путь к файлу для хранения данных") // тут, если будет запущен без флага -file-storage-path, то по умолчанию будет исп. файл storage.json (он будет создаваться в той папке откуда запущен автотест/сервер, обычно автотест запускаю из корня /yaGo)

	// Парсим флаги (обязательно!)
	flag.Parse()

	// 2. Проверяем переменные окружения (они перезаписывают флаги)
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.ServerAddress = envAddr
	}

	if envBase := os.Getenv("BASE_URL"); envBase != "" {
		cfg.BaseURL = envBase
	}

	if envPath := os.Getenv("FILE_STORAGE_PATH"); envPath != "" {
		cfg.FileStoragePath = envPath
	}
	return cfg
}

/*
Разберем всё что выше:
1. в этой команде cfg := &Config{}:
- Config{} - это создание совершенно нового, пустого экземпляра (объекта) по шаблону который объявляем вначале type Config struct
- и в амперсанд (&) перед вновь созданым объектом Config{} - берет адрес этого объекта в памяти, и присваиваем этот адрес пемен. cfg
- т.е. cfg содержит адрес нового экземпляра объекта по шаблону структуры Config. И тип у перемен. cfg - это указатель на структуру Config (*Config). Внутри неё (по этому адресу) лежат пустые поля ServerAddress и BaseURL

2. в flag.StringVar(&cfg.ServerAddress, ...):
В данном случае в первом аргументе функции передается адрес конкретного поля структуры
- cfg — это адрес структуры.
- cfg.ServerAddress — это само поле (строка) внутри структуры.
- &cfg.ServerAddress — это адрес именно этого поля в памяти.
И фукнции flag.StringVar() нужен адрес полня, чтобы она могла напрямую записать туда значение.

Как это можно представить в памяти:
1. Config{}: Ты создал шкаф, в котором две полки: на одной лежит строка ServerAddress, на другой BaseURL.
2. &Config{}: Ты взял номер этого шкафа (например, шкаф №101) и записал его в переменную cfg.
3. cfg.ServerAddress: Ты говоришь: «Зайди в шкаф №101 и возьми то, что лежит на полке ServerAddress».
4. &cfg.ServerAddress: Ты говоришь: «Зайди в шкаф №101, найди полку ServerAddress и передай мне точный адрес этой полки».

flag.StringVar берёт этот адрес полки и кладёт туда значение, когда парсит командную строку.

Резюмируя:
1. Config — это чертёж (шаблон).
2. Config{} — это новый объект, созданный по чертежу.
3. cfg := &Config{} — cfg хранит адрес этого нового объекта.
4. &cfg.ServerAddress — это адрес конкретного поля внутри объекта, чтобы функция flag могла записать туда данные.

*/
