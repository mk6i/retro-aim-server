CREATE TABLE chatRoom
(
    cookie   TEXT PRIMARY KEY,
    exchange INTEGER,
    name     TEXT,
    created  TIMESTAMP,
    creator  TEXT,
    UNIQUE (exchange, name)
);
