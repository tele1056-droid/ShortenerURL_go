package shortener

import "sync"

var (
	mu sync.Mutex // используем один мьютекс на весь пакет (т.е. для Shorten() и Get()).  Т.к. защищаем один ресурс (мапу urls и счётчик counter), то один замок(мьютекс) для одного ресурса.

	//вариант с глобальной map: Ключ - короткий код, Значение - длинный URL
	urls = make(map[string]string)

	//объявляем глобальный счётчик, по дефолту значение ноль, это значение мы будем использовать при вызове toBase62() в хендлере, и в хендлере увеличивать ++
	counter int64
)


//переводим число в строку Base62
func toBase62(n int64) string {
	if n == 0 {
		return "0"
	}
	
	const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := ""
	for n > 0 {
		remainder := n % 62
		result = string(base62Chars[remainder]) + result
		n = n / 62
	}
	return result
}

func Shorten(longURL string) string {
	//тут можно добавить блок кода для "мьютекса" (защиты от гонок)
	mu.Lock()
	defer mu.Unlock()
	

	counter++
	code := toBase62(counter)
	urls[code] = longURL
	return code
}

// Get ищет код в карте и возвращает URL + флаг «найдено»
	// и тут в любом случае нужно использовать мьютекс чтобы избежать гонки, т.к. HTTP-сервер Go создаёт горутину для каждого входящего запроса автоматически
func Get(code string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	longURL, ok := urls[code]
	return longURL, ok
}

/*
Функция (func Get()) возвращает два значения:
	- string — сам URL (или пустую строку, если не найден).
	- bool — флаг: true, если код найден, false, если нет.
Это стандартный паттерн в Go для проверки наличия ключа в карте:
	value, ok := myMap[key]
		if !ok {
    	// ключа нет
	}
*/