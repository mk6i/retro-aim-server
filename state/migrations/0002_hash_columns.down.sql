CREATE TEMPORARY TABLE user_backup
(
    screenName VARCHAR(16) PRIMARY KEY,
    authKey    TEXT,
    passHash   TEXT
);

INSERT INTO user_backup (screenName, authKey, passHash)
SELECT screenName, authKey, strongMD5Pass
FROM user;

DROP TABLE user;

ALTER TABLE user_backup
    RENAME TO user;
