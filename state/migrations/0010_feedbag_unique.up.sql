CREATE TABLE feedbag_new
(
    screenName   VARCHAR(16),
    groupID      INTEGER,
    itemID       INTEGER,
    classID      INTEGER,
    name         TEXT,
    attributes   BLOB,
    lastModified INTEGER,
    UNIQUE (screenName, groupID, itemID),
    UNIQUE (screenName, groupID, classID, name)
);

INSERT INTO feedbag_new
SELECT *
FROM feedbag;

DROP TABLE feedbag;

ALTER TABLE feedbag_new
    RENAME TO feedbag;