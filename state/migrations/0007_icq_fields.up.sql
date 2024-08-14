ALTER TABLE users
    ADD COLUMN isICQ BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users
    ADD COLUMN icq_affiliations_currentCode1 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_affiliations_currentCode2 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_affiliations_currentCode3 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_affiliations_currentKeyword1 TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_affiliations_currentKeyword2 TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_affiliations_currentKeyword3 TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_affiliations_pastCode1 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_affiliations_pastCode2 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_affiliations_pastCode3 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_affiliations_pastKeyword1 TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_affiliations_pastKeyword2 TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_affiliations_pastKeyword3 TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_address TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_cellPhone TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_city TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_countryCode INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_basicInfo_emailAddress TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_fax TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_firstName TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_gmtOffset INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_basicInfo_lastName TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_nickName TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_phone TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_publishEmail BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users
    ADD COLUMN icq_basicInfo_state TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_basicInfo_zipCode TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_interests_code1 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_interests_code2 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_interests_code3 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_interests_code4 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_interests_keyword1 TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_interests_keyword2 TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_interests_keyword3 TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_interests_keyword4 TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_moreInfo_birthDay INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_moreInfo_birthMonth INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_moreInfo_birthYear INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_moreInfo_gender INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_moreInfo_homePageAddr TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_moreInfo_lang1 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_moreInfo_lang2 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_moreInfo_lang3 INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_notes TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_permissions_authRequired BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users
    ADD COLUMN icq_workInfo_address TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_workInfo_city TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_workInfo_company TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_workInfo_countryCode INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_workInfo_department TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_workInfo_fax TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_workInfo_occupationCode INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users
    ADD COLUMN icq_workInfo_phone TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_workInfo_position TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_workInfo_state TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_workInfo_webPage TEXT NOT NULL DEFAULT '';
ALTER TABLE users
    ADD COLUMN icq_workInfo_zipCode TEXT NOT NULL DEFAULT '';