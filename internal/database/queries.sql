-- name: FetchOne :one
SELECT *
FROM sources
WHERE last_visited < sqlc.arg(unix_time_until)
  AND retries < sqlc.arg(max_retries)
ORDER BY retries
LIMIT 1;

-- name: UpdateLastVisited :exec
UPDATE sources
SET last_visited = sqlc.arg(last_visited)
WHERE id = sqlc.arg(id);

-- name: ResetRetries :exec
UPDATE sources
SET retries = 0;

-- name: UpdateRetries :exec
UPDATE sources
SET retries = retries + 1
WHERE id = sqlc.arg(id);

-- name: UpsertSource :exec
INSERT INTO sources (name, url, config)
VALUES (sqlc.arg(name),
        sqlc.arg(url),
        sqlc.arg(config))
ON CONFLICT (url)
    DO UPDATE
    SET config = excluded.config;

-- name: UpsertArticle :exec
INSERT INTO articles (source_id, title, url, added)
VALUES (sqlc.arg(source_id),
        sqlc.arg(title),
        sqlc.arg(url),
        sqlc.arg(added))
ON CONFLICT (url)
    DO NOTHING;

-- name: LastTimestamp :one
SELECT timestamp
FROM timestamp
ORDER BY timestamp DESC
LIMIT 1;

-- name: SetLastTimestamp :exec
INSERT INTO timestamp (timestamp)
VALUES (sqlc.arg(timestamp));
COMMIT;

-- name: SelectUnsent :one
SELECT a.id,
       a.title,
       a.url,
       s.name as source_name
FROM articles as a
JOIN sources s on s.id = a.source_id
WHERE sent = 0
LIMIT 1;

-- name: MarkArticleSent :exec
UPDATE articles
SET sent = 1
WHERE id = sqlc.arg(id);
