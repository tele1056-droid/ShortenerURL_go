package storage

type Repository interface {
	Save(code, longURL, userID string) error
	Get(code string) (string, bool)
	GetAllByUser(userID string) (map[string]string, error) // новый метод
}
