ALTER TABLE user
    RENAME COLUMN passHash TO strongMD5Pass;

ALTER TABLE user
    ADD COLUMN weakMD5Pass TEXT;

-- The cleartext passwords don't exist, so it's not possible to create weak MD5
-- hashes. Fill in the values with a placeholder value. The administrator will
-- require everyone to reset their passwords if they want to log in with AIM
-- 3.5 thru AIM4.7
UPDATE user
SET weakMD5Pass = strongMD5Pass;