CREATE TABLE user
(
    screenName    VARCHAR(16) PRIMARY KEY,
    authKey       TEXT,
    strongMD5Pass TEXT,
    weakMD5Pass   TEXT
);

INSERT INTO user
SELECT displayScreenName,
       authKey,
       strongMD5Pass,
       weakMD5Pass
FROM users;

DROP TABLE users;