ALTER TABLE feedbag
	ADD COLUMN pdMode int DEFAULT 0;

UPDATE feedbag
SET pdMode =
		CASE
			WHEN instr(hex(attributes), '00CA000101') > 0 THEN 1
			WHEN instr(hex(attributes), '00CA000102') > 0 THEN 2
			WHEN instr(hex(attributes), '00CA000103') > 0 THEN 3
			WHEN instr(hex(attributes), '00CA000104') > 0 THEN 4
			WHEN instr(hex(attributes), '00CA000105') > 0 THEN 5
			END
WHERE classID = 4;

CREATE TABLE buddyListMode
(
	screenName       VARCHAR(16),
	clientSidePDMode INTEGER DEFAULT 0,
	useFeedbag       BOOLEAN DEFAULT false,
	PRIMARY KEY (screenName)
);

CREATE TABLE clientSideBuddyList
(
	me       VARCHAR(16),
	them     VARCHAR(16),
	isBuddy  BOOLEAN DEFAULT false,
	isPermit BOOLEAN DEFAULT false,
	isDeny   BOOLEAN DEFAULT false,
	PRIMARY KEY (me, them)
);