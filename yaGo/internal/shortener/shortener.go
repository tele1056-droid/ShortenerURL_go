package shortener

import "sync"

var (
	mu sync.Mutex // используем один мьютекс на весь пакет (т.е. для Shorten() и Get()).  Т.к. защищаем один ресурс (мапу urls и счётчик counter), то один замок(мьютекс) для одного ресурса.

	//объявляем глобальный счётчик, по дефолту значение ноль, это значение мы будем использовать при вызове toBase62() в хендлере, и в хендлере увеличивать ++
	counter int64
)

// переводим число в строку Base62
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

//в этой версии, Shortener() принимает длинный URL и возвращает сгенерированный код, но не сохраняет его - сохраняет теперь repo.Save()

// в этой версии убираем из сигнатуры Shorten() аргумент и в пакете handlers тоже меняем вызов func Shorten()
func Shorten() string {
	//тут можно добавить блок кода для "мьютекса" (защиты от гонок)
	mu.Lock()
	defer mu.Unlock()

	counter++
	code := toBase62(counter)
	return code
}

/*
shortener больше не знает о карте и не сохраняет в неё. Это ответственность storage.

Про счётчик counter int64. Счётчик оставляем тут в пакете shortener, как глобальный, но минусы в том что счётчик - глобальное состояние, при перезапуске сбрасывается.

И есть вариант перенести счётчик в storage
*/
