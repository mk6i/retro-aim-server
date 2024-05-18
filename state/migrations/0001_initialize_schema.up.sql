CREATE TABLE IF NOT EXISTS "user"
(
    "screenName" VARCHAR(16) PRIMARY KEY,
    "authKey"    TEXT,
    "passHash"   TEXT
);

CREATE TABLE IF NOT EXISTS "feedbag"
(
    "screenName"   VARCHAR(16),
    "groupID"      INTEGER,
    "itemID"       INTEGER,
    "classID"      INTEGER,
    "name"         TEXT,
    "attributes"   BYTEA,
    "lastModified" INTEGER,
    UNIQUE ("screenName", "groupID", "itemID")
);

CREATE TABLE IF NOT EXISTS "profile"
(
    "screenName" VARCHAR(16) PRIMARY KEY,
    "body"       TEXT
);

CREATE TABLE IF NOT EXISTS "bartItem"
(
    "hash" CHAR(16) PRIMARY KEY,
    "body" BYTEA
);