package database

import (
	"context"
	"database/sql"

	"github.com/denisdubovitskiy/feedparser/internal/config"
	"github.com/denisdubovitskiy/feedparser/internal/database/queries"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Service struct {
	queries *queries.Queries
}

func NewService(db DBTX) *Service {
	return &Service{
		queries: queries.New(db),
	}
}

type Source struct {
	ID          int64
	URL         string
	Name        string
	Config      config.SourceConfig
	LastVisited int64
	Retries     int64
}

type SaveArticleParams = queries.UpsertArticleParams

func (s *Service) SaveArticle(ctx context.Context, params SaveArticleParams) error {
	return s.queries.UpsertArticle(ctx, params)
}

type Article struct {
	Title string
	URL   string
}

func (s *Service) UnsentArticles(ctx context.Context) ([]Article, error) {
	rows, err := s.queries.SelectUnsent(ctx)
	if err != nil {
		return nil, err
	}

	articles := make([]Article, len(rows))
	for i, row := range rows {
		articles[i] = Article{
			Title: row.Title,
			URL:   row.Url,
		}
	}

	return articles, nil
}

func (s *Service) FetchOne(ctx context.Context, unixTimeUntil, maxRetries int64) (*Source, error) {
	source, err := s.queries.FetchOne(ctx, queries.FetchOneParams{
		UnixTimeUntil: unixTimeUntil,
		MaxRetries:    maxRetries,
	})
	if err != nil {
		return nil, err
	}

	conf, err := config.ParseSourceConfig([]byte(source.Config))
	if err != nil {
		return nil, err
	}

	return &Source{
		ID:          source.ID,
		URL:         source.Url,
		Name:        source.Name,
		Config:      conf,
		LastVisited: source.LastVisited,
		Retries:     source.Retries,
	}, err
}

func (s *Service) UpsertSource(ctx context.Context, name, url, config string) error {
	return s.queries.UpsertSource(ctx, queries.UpsertSourceParams{
		Name:   name,
		Url:    url,
		Config: config,
	})
}

func (s *Service) ResetRetries(ctx context.Context) error {
	return s.queries.ResetRetries(ctx)
}
func (s *Service) UpdateRetries(ctx context.Context, id int64) error {
	return s.queries.UpdateRetries(ctx, id)
}

func (s *Service) UpdateLastVisited(ctx context.Context, id, unixTimeUntil int64) error {
	return s.queries.UpdateLastVisited(ctx, queries.UpdateLastVisitedParams{
		ID:          id,
		LastVisited: unixTimeUntil,
	})
}

func (s *Service) LastTimestamp(ctx context.Context) (int64, error) {
	return s.queries.LastTimestamp(ctx)
}

func (s *Service) SetLastTimestamp(ctx context.Context, timestamp int64) error {
	return s.queries.SetLastTimestamp(ctx, timestamp)
}
