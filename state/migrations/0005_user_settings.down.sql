CREATE TABLE users_backup
(
    identScreenName   VARCHAR(16) PRIMARY KEY,
    displayScreenName TEXT,
    authKey           TEXT,
    strongMD5Pass     TEXT,
    weakMD5Pass       TEXT
);

INSERT INTO users_backup (identScreenName, displayScreenName, authKey, strongMD5Pass, weakMD5Pass)
SELECT identScreenName, displayScreenName, authKey, strongMD5Pass, weakMD5Pass
FROM users;

DROP TABLE users;

ALTER TABLE users_backup
    RENAME TO users;
