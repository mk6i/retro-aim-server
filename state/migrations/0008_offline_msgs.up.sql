CREATE TABLE offlineMessage
(
    sender      VARCHAR(16) NOT NULL,
    recipient   VARCHAR(16) NOT NULL,
    message     BLOB NOT NULL,
    sent        TIMESTAMP NOT NULL
);