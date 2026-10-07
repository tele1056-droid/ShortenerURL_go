package storage

type Repository interface {
	Save(code, longURL string) error
	Get(code string) (string, bool)
}