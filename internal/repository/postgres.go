package repository

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/siva-kannan3/go-url-shortener/internal/domain"
)

var (
	ErrIDAlreadyExists = errors.New("id already exists")
	ErrURLNotFound     = errors.New("url not found")
)

type URLRepository interface {
	Get(id string) (domain.ShortenedURL, error)
	Create(ShortenedURL domain.ShortenedURL) (domain.ShortenedURL, error)
}

// Compile-time check to ensure PostgresRepository implements URLRepository.
// If later someone changes the interface: the project will fail to compile until PostgresRepository implement the change
var _ URLRepository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (repository *PostgresRepository) Create(shortenedUrl domain.ShortenedURL) (domain.ShortenedURL, error) {
	_, err := repository.db.Exec(`INSERT INTO urls (short_id, url) VALUES ($1, $2)`, shortenedUrl.ID, shortenedUrl.Url)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ShortenedURL{}, ErrIDAlreadyExists
		}

		return domain.ShortenedURL{}, err
	}

	return shortenedUrl, nil
}

func (repository *PostgresRepository) Get(id string) (domain.ShortenedURL, error) {
	var shortenedURL domain.ShortenedURL

	err := repository.db.QueryRow(`SELECT * FROM urls WHERE short_id = $1`, id).Scan(&shortenedURL.ID, &shortenedURL.Url)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ShortenedURL{}, ErrURLNotFound
		}

		return domain.ShortenedURL{}, err
	}

	return shortenedURL, err
}
