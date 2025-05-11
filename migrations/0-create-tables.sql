CREATE TABLE IF NOT EXISTS hashes
(
    id           INTEGER PRIMARY KEY,
    url          TEXT NOT NULL,
    content_hash TEXT UNIQUE,
    created_at    DATETIME,
    updated_at    DATETIME
);

CREATE TABLE IF NOT EXISTS phrases
(
    id          INTEGER PRIMARY KEY,
    author      TEXT NOT NULL,
    phrase      TEXT NOT NULL,
    phrase_hash TEXT UNIQUE,
    created_at   DATETIME,
    updated_at   DATETIME,
    hash_id     TEXT NOT NULL
);
