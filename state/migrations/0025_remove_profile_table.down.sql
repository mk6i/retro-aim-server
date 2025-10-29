-- Recreate the profile table
CREATE TABLE IF NOT EXISTS profile
(
    screenName VARCHAR(16) PRIMARY KEY,
    body       TEXT
);

