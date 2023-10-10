-- DROP TABLE IF EXISTS sources;

CREATE TABLE IF NOT EXISTS sources
(
    id           INTEGER PRIMARY KEY NOT NULL        DEFAULT 0,
    url          TEXT                NOT NULL UNIQUE DEFAULT '',
    name         TEXT                NOT NULL UNIQUE DEFAULT '',
    config       TEXT                NOT NULL        DEFAULT '',
    last_visited INTEGER             NOT NULL        DEFAULT 0,
    retries      INTEGER             NOT NULL        DEFAULT 0
);

CREATE TABLE IF NOT EXISTS articles
(
    id        INTEGER PRIMARY KEY NOT NULL        DEFAULT 0,
    source_id INTEGER             NOT NULL        DEFAULT 0,
    title     TEXT                NOT NULL        DEFAULT '',
    url       TEXT                NOT NULL UNIQUE DEFAULT '',
    sent      INTEGER             NOT NULL        DEFAULT 0,
    added     INTEGER             NOT NULL        DEFAULT 0
);

CREATE TABLE IF NOT EXISTS timestamp (
    timestamp INTEGER NOT NULL default 0
);
