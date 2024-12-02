CREATE TABLE feedbag_new
(
	screenName   VARCHAR(16),
	groupID      INTEGER,
	itemID       INTEGER,
	classID      INTEGER,
	name         TEXT,
	attributes   BLOB,
	lastModified INTEGER,
	UNIQUE (screenName, groupID, itemID)
);

INSERT INTO feedbag_new (screenName, groupID, itemID, classID, name, attributes, lastModified)
SELECT screenName, groupID, itemID, classID, name, attributes, lastModified
FROM feedbag;

DROP TABLE feedbag;

ALTER TABLE feedbag_new
	RENAME TO feedbag;

DROP TABLE buddyListMode;
DROP TABLE clientSideBuddyList;
