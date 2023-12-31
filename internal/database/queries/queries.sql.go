// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: queries.sql

package queries

import (
	"context"
)

const fetchOne = `-- name: FetchOne :one
SELECT id, url, name, config, last_visited, retries
FROM sources
WHERE last_visited < ?1
  AND retries < ?2
ORDER BY retries
LIMIT 1
`

type FetchOneParams struct {
	UnixTimeUntil int64
	MaxRetries    int64
}

func (q *Queries) FetchOne(ctx context.Context, arg FetchOneParams) (Source, error) {
	row := q.db.QueryRowContext(ctx, fetchOne, arg.UnixTimeUntil, arg.MaxRetries)
	var i Source
	err := row.Scan(
		&i.ID,
		&i.Url,
		&i.Name,
		&i.Config,
		&i.LastVisited,
		&i.Retries,
	)
	return i, err
}

const lastTimestamp = `-- name: LastTimestamp :one
SELECT timestamp
FROM timestamp
ORDER BY timestamp DESC
LIMIT 1
`

func (q *Queries) LastTimestamp(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, lastTimestamp)
	var timestamp int64
	err := row.Scan(&timestamp)
	return timestamp, err
}

const markArticleSent = `-- name: MarkArticleSent :exec
UPDATE articles
SET sent = 1
WHERE id = ?1
`

func (q *Queries) MarkArticleSent(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, markArticleSent, id)
	return err
}

const resetRetries = `-- name: ResetRetries :exec
UPDATE sources
SET retries = 0
`

func (q *Queries) ResetRetries(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, resetRetries)
	return err
}

const selectUnsent = `-- name: SelectUnsent :one
COMMIT;

SELECT a.id,
       a.title,
       a.url,
       s.name as source_name,
       s.config as config
FROM articles as a
JOIN sources s on s.id = a.source_id
WHERE sent = 0
LIMIT 1
`

type SelectUnsentRow struct {
	ID         int64
	Title      string
	Url        string
	SourceName string
	Config     string
}

func (q *Queries) SelectUnsent(ctx context.Context) (SelectUnsentRow, error) {
	row := q.db.QueryRowContext(ctx, selectUnsent)
	var i SelectUnsentRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Url,
		&i.SourceName,
		&i.Config,
	)
	return i, err
}

const setLastTimestamp = `-- name: SetLastTimestamp :exec
INSERT INTO timestamp (timestamp)
VALUES (?1)
`

func (q *Queries) SetLastTimestamp(ctx context.Context, timestamp int64) error {
	_, err := q.db.ExecContext(ctx, setLastTimestamp, timestamp)
	return err
}

const updateLastVisited = `-- name: UpdateLastVisited :exec
UPDATE sources
SET last_visited = ?1
WHERE id = ?2
`

type UpdateLastVisitedParams struct {
	LastVisited int64
	ID          int64
}

func (q *Queries) UpdateLastVisited(ctx context.Context, arg UpdateLastVisitedParams) error {
	_, err := q.db.ExecContext(ctx, updateLastVisited, arg.LastVisited, arg.ID)
	return err
}

const updateRetries = `-- name: UpdateRetries :exec
UPDATE sources
SET retries = retries + 1
WHERE id = ?1
`

func (q *Queries) UpdateRetries(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, updateRetries, id)
	return err
}

const upsertArticle = `-- name: UpsertArticle :exec
INSERT INTO articles (source_id, title, url, added)
VALUES (?1,
        ?2,
        ?3,
        ?4)
ON CONFLICT (url)
    DO NOTHING
`

type UpsertArticleParams struct {
	SourceID int64
	Title    string
	Url      string
	Added    int64
}

func (q *Queries) UpsertArticle(ctx context.Context, arg UpsertArticleParams) error {
	_, err := q.db.ExecContext(ctx, upsertArticle,
		arg.SourceID,
		arg.Title,
		arg.Url,
		arg.Added,
	)
	return err
}

const upsertSource = `-- name: UpsertSource :exec
INSERT INTO sources (name, url, config)
VALUES (?1,
        ?2,
        ?3)
ON CONFLICT (url)
    DO UPDATE
    SET config = excluded.config
`

type UpsertSourceParams struct {
	Name   string
	Url    string
	Config string
}

func (q *Queries) UpsertSource(ctx context.Context, arg UpsertSourceParams) error {
	_, err := q.db.ExecContext(ctx, upsertSource, arg.Name, arg.Url, arg.Config)
	return err
}
