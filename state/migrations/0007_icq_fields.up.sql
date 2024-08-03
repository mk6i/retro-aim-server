ALTER TABLE users
    ADD COLUMN firstName TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN lastName TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN nickName TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN authReq BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users
    ADD COLUMN gender INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN homeCity TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN homeState TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN homePhone TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN homeFax TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN homeAddress TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN cellPhone TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN zipCode TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN countryCode INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN gmtOffset INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN publishEmail BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users
    ADD COLUMN workCity TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN workState TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN workPhone TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN workFax TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN workAddress TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN workZIP TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN workCountryCode INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN company TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN department TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN position TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN occupationCode INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN workWebPage TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN homePageAddr TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN birthYear INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN birthMonth INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN birthDay INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN lang1 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN lang2 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN lang3 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN notes TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN interest1Code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN interest1Keyword TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN interest2Code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN interest2Keyword TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN interest3Code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN interest3Keyword TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN interest4Code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN interest4Keyword TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN pastAffiliations1Code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN pastAffiliations1Keyword TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN pastAffiliations2Code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN pastAffiliations2Keyword TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN pastAffiliations3Code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN pastAffiliations3Keyword TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN affiliations1Code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN affiliations1Keyword TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN affiliations2Code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN affiliations2Keyword TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN affiliations3Code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN affiliations3Keyword TEXT NOT NULL DEFAULT '';

-- INSERT INTO users (
--     identScreenName, displayScreenName, authKey, strongMD5Pass, weakMD5Pass,
--     confirmStatus, emailAddress, regStatus, firstName, lastName,
--     authReq, gender, birthday, homeCity, homeState,
--     homePhone, homeFax, homeAddress, cellPhone, zipCode,
--     countryCode, gmtOffset, publishEmail
-- ) VALUES
--       (
--           'user1', 'User One', 'authKey1', 'strongPass1', 'weakPass1',
--           TRUE, 'user1@example.com', 3, 'John', 'Doe',
--           FALSE, 1, 19900101, 'New York', 'NY',
--           '123-456-7890', '', '123 Main St', '321-654-0987', '10001',
--           1, -5, TRUE
--       ),
--       (
--           'user2', 'User Two', 'authKey2', 'strongPass2', 'weakPass2',
--           FALSE, 'user2@example.com', 3, 'Jane', 'Smith',
--           TRUE, 2, 19850515, 'Los Angeles', 'CA',
--           '234-567-8901', '', '456 Elm St', '432-765-1098', '90001',
--           1, -8, FALSE
--       ),
--       (
--           'user3', 'User Three', 'authKey3', 'strongPass3', 'weakPass3',
--           TRUE, 'user3@example.com', 3, 'Alice', 'Johnson',
--           FALSE, 1, 19921010, 'Chicago', 'IL',
--           '345-678-9012', '', '789 Oak St', '543-876-2109', '60601',
--           1, -6, TRUE
--       ),
--       (
--           'user4', 'User Four', 'authKey4', 'strongPass4', 'weakPass4',
--           FALSE, 'user4@example.com', 3, 'Bob', 'Brown',
--           TRUE, 1, 19781225, 'Houston', 'TX',
--           '456-789-0123', '', '101 Pine St', '654-987-3210', '77001',
--           1, -6, FALSE
--       ),
--       (
--           'user5', 'User Five', 'authKey5', 'strongPass5', 'weakPass5',
--           TRUE, 'user5@example.com', 3, 'Charlie', 'Davis',
--           FALSE, 2, 20000707, 'Miami', 'FL',
--           '567-890-1234', '', '202 Cedar St', '765-098-4321', '33101',
--           1, -5, TRUE
--       );
