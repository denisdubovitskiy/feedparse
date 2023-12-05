package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

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
	db      *sql.DB
	mu      sync.Mutex
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:      db,
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

func (s Source) String() string {
	return fmt.Sprintf("Source(id=%d, name=%s)", s.ID, s.Name)
}

type SaveArticleParams = queries.UpsertArticleParams

func (s *Service) SaveArticle(ctx context.Context, params SaveArticleParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queries.UpsertArticle(ctx, params)
}

type Article struct {
	Title  string
	URL    string
	Source string
	Tags   []string
}

func (a Article) String() string {
	return fmt.Sprintf("Article(title=%s, url=%s)", a.Title, a.URL)
}

func (s *Service) SelectUnsent(ctx context.Context, f func(a Article) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	q := s.queries.WithTx(tx)

	article, err := q.SelectUnsent(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	var conf config.SourceConfig
	if err := json.Unmarshal([]byte(article.Config), &conf); err != nil {
		return err
	}

	fnErr := f(Article{
		Title:  article.Title,
		URL:    article.Url,
		Source: article.SourceName,
		Tags:   conf.Tags,
	})
	if fnErr != nil {
		_ = tx.Rollback()
		return fnErr
	}

	if sendErr := q.MarkArticleSent(ctx, article.ID); sendErr != nil {
		_ = tx.Rollback()
		return sendErr
	}

	_ = tx.Commit()

	return nil
}

func (s *Service) FetchOne(ctx context.Context, unixTimeUntil, maxRetries int64) (*Source, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queries.UpsertSource(ctx, queries.UpsertSourceParams{
		Name:   name,
		Url:    url,
		Config: config,
	})
}

func (s *Service) ResetRetries(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queries.ResetRetries(ctx)
}
func (s *Service) UpdateRetries(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queries.UpdateRetries(ctx, id)
}

func (s *Service) UpdateLastVisited(ctx context.Context, id, unixTimeUntil int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queries.UpdateLastVisited(ctx, queries.UpdateLastVisitedParams{
		ID:          id,
		LastVisited: unixTimeUntil,
	})
}

func (s *Service) LastTimestamp(ctx context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queries.LastTimestamp(ctx)
}

func (s *Service) SetLastTimestamp(ctx context.Context, timestamp int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queries.SetLastTimestamp(ctx, timestamp)
}
