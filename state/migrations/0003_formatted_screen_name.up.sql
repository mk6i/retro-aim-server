-- call the new table "users" since it's not a reserved word that we'll need to
-- deal with for postgres in the future.
CREATE TABLE users
(
    identScreenName   VARCHAR(16) PRIMARY KEY,
    displayScreenName TEXT,
    authKey           TEXT,
    strongMD5Pass     TEXT,
    weakMD5Pass       TEXT
);

INSERT INTO users
SELECT LOWER(REPLACE(screenName, ' ', '')),
       screenName,
       authKey,
       strongMD5Pass,
       weakMD5Pass
FROM user;

DROP TABLE user;

UPDATE feedbag
SET name = LOWER(REPLACE(name, ' ', ''))
WHERE classID IN (0, 2, 3);

UPDATE feedbag
SET screenName = LOWER(REPLACE(screenName, ' ', ''));