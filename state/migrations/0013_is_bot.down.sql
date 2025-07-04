ALTER TABLE users
    RENAME TO users_old;

CREATE TABLE users
(
    identScreenName                  VARCHAR(16) PRIMARY KEY,
    displayScreenName                TEXT,
    authKey                          TEXT,
    strongMD5Pass                    TEXT,
    weakMD5Pass                      TEXT,
    confirmStatus                    BOOLEAN               DEFAULT FALSE,
    emailAddress                     VARCHAR(320) NOT NULL DEFAULT '',
    regStatus                        INT          NOT NULL DEFAULT 3,
    isICQ                            BOOLEAN      NOT NULL DEFAULT false,
    aim_firstName                    TEXT         NOT NULL DEFAULT '',
    aim_lastName                     TEXT         NOT NULL DEFAULT '',
    aim_middleName                   TEXT         NOT NULL DEFAULT '',
    aim_maidenName                   TEXT         NOT NULL DEFAULT '',
    aim_country                      TEXT         NOT NULL DEFAULT '',
    aim_state                        TEXT         NOT NULL DEFAULT '',
    aim_city                         TEXT         NOT NULL DEFAULT '',
    aim_nickName                     TEXT         NOT NULL DEFAULT '',
    aim_zipCode                      TEXT         NOT NULL DEFAULT '',
    aim_address                      TEXT         NOT NULL DEFAULT '',
    aim_keyword1                     INTEGER,
    aim_keyword2                     INTEGER,
    aim_keyword3                     INTEGER,
    aim_keyword4                     INTEGER,
    aim_keyword5                     INTEGER,
    icq_affiliations_currentCode1    INTEGER      NOT NULL DEFAULT 0,
    icq_affiliations_currentCode2    INTEGER      NOT NULL DEFAULT 0,
    icq_affiliations_currentCode3    INTEGER      NOT NULL DEFAULT 0,
    icq_affiliations_currentKeyword1 TEXT         NOT NULL DEFAULT '',
    icq_affiliations_currentKeyword2 TEXT         NOT NULL DEFAULT '',
    icq_affiliations_currentKeyword3 TEXT         NOT NULL DEFAULT '',
    icq_affiliations_pastCode1       INTEGER      NOT NULL DEFAULT 0,
    icq_affiliations_pastCode2       INTEGER      NOT NULL DEFAULT 0,
    icq_affiliations_pastCode3       INTEGER      NOT NULL DEFAULT 0,
    icq_affiliations_pastKeyword1    TEXT         NOT NULL DEFAULT '',
    icq_affiliations_pastKeyword2    TEXT         NOT NULL DEFAULT '',
    icq_affiliations_pastKeyword3    TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_address            TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_cellPhone          TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_city               TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_countryCode        INTEGER      NOT NULL DEFAULT 0,
    icq_basicInfo_emailAddress       TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_fax                TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_firstName          TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_gmtOffset          INTEGER      NOT NULL DEFAULT 0,
    icq_basicInfo_lastName           TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_nickName           TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_phone              TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_publishEmail       BOOLEAN      NOT NULL DEFAULT false,
    icq_basicInfo_state              TEXT         NOT NULL DEFAULT '',
    icq_basicInfo_zipCode            TEXT         NOT NULL DEFAULT '',
    icq_interests_code1              INTEGER      NOT NULL DEFAULT 0,
    icq_interests_code2              INTEGER      NOT NULL DEFAULT 0,
    icq_interests_code3              INTEGER      NOT NULL DEFAULT 0,
    icq_interests_code4              INTEGER      NOT NULL DEFAULT 0,
    icq_interests_keyword1           TEXT         NOT NULL DEFAULT '',
    icq_interests_keyword2           TEXT         NOT NULL DEFAULT '',
    icq_interests_keyword3           TEXT         NOT NULL DEFAULT '',
    icq_interests_keyword4           TEXT         NOT NULL DEFAULT '',
    icq_moreInfo_birthDay            INTEGER      NOT NULL DEFAULT 0,
    icq_moreInfo_birthMonth          INTEGER      NOT NULL DEFAULT 0,
    icq_moreInfo_birthYear           INTEGER      NOT NULL DEFAULT 0,
    icq_moreInfo_gender              INTEGER      NOT NULL DEFAULT 0,
    icq_moreInfo_homePageAddr        TEXT         NOT NULL DEFAULT '',
    icq_moreInfo_lang1               INTEGER      NOT NULL DEFAULT 0,
    icq_moreInfo_lang2               INTEGER      NOT NULL DEFAULT 0,
    icq_moreInfo_lang3               INTEGER      NOT NULL DEFAULT 0,
    icq_notes                        TEXT         NOT NULL DEFAULT '',
    icq_permissions_authRequired     BOOLEAN      NOT NULL DEFAULT false,
    icq_workInfo_address             TEXT         NOT NULL DEFAULT '',
    icq_workInfo_city                TEXT         NOT NULL DEFAULT '',
    icq_workInfo_company             TEXT         NOT NULL DEFAULT '',
    icq_workInfo_countryCode         INTEGER      NOT NULL DEFAULT 0,
    icq_workInfo_department          TEXT         NOT NULL DEFAULT '',
    icq_workInfo_fax                 TEXT         NOT NULL DEFAULT '',
    icq_workInfo_occupationCode      INTEGER      NOT NULL DEFAULT 0,
    icq_workInfo_phone               TEXT         NOT NULL DEFAULT '',
    icq_workInfo_position            TEXT         NOT NULL DEFAULT '',
    icq_workInfo_state               TEXT         NOT NULL DEFAULT '',
    icq_workInfo_webPage             TEXT         NOT NULL DEFAULT '',
    icq_workInfo_zipCode             TEXT         NOT NULL DEFAULT '',
    tocConfig                        TEXT         NOT NULL DEFAULT '',
    suspendedStatus                  INT          NOT NULL DEFAULT 0,

    FOREIGN KEY (aim_keyword1) REFERENCES aimKeyword (id),
    FOREIGN KEY (aim_keyword2) REFERENCES aimKeyword (id),
    FOREIGN KEY (aim_keyword3) REFERENCES aimKeyword (id),
    FOREIGN KEY (aim_keyword4) REFERENCES aimKeyword (id),
    FOREIGN KEY (aim_keyword5) REFERENCES aimKeyword (id)
);

INSERT INTO users (identScreenName,
                   displayScreenName,
                   authKey,
                   strongMD5Pass,
                   weakMD5Pass,
                   confirmStatus,
                   emailAddress,
                   regStatus,
                   isICQ,
                   aim_firstName,
                   aim_lastName,
                   aim_middleName,
                   aim_maidenName,
                   aim_country,
                   aim_state,
                   aim_city,
                   aim_nickName,
                   aim_zipCode,
                   aim_address,
                   aim_keyword1,
                   aim_keyword2,
                   aim_keyword3,
                   aim_keyword4,
                   aim_keyword5,
                   icq_affiliations_currentCode1,
                   icq_affiliations_currentCode2,
                   icq_affiliations_currentCode3,
                   icq_affiliations_currentKeyword1,
                   icq_affiliations_currentKeyword2,
                   icq_affiliations_currentKeyword3,
                   icq_affiliations_pastCode1,
                   icq_affiliations_pastCode2,
                   icq_affiliations_pastCode3,
                   icq_affiliations_pastKeyword1,
                   icq_affiliations_pastKeyword2,
                   icq_affiliations_pastKeyword3,
                   icq_basicInfo_address,
                   icq_basicInfo_cellPhone,
                   icq_basicInfo_city,
                   icq_basicInfo_countryCode,
                   icq_basicInfo_emailAddress,
                   icq_basicInfo_fax,
                   icq_basicInfo_firstName,
                   icq_basicInfo_gmtOffset,
                   icq_basicInfo_lastName,
                   icq_basicInfo_nickName,
                   icq_basicInfo_phone,
                   icq_basicInfo_publishEmail,
                   icq_basicInfo_state,
                   icq_basicInfo_zipCode,
                   icq_interests_code1,
                   icq_interests_code2,
                   icq_interests_code3,
                   icq_interests_code4,
                   icq_interests_keyword1,
                   icq_interests_keyword2,
                   icq_interests_keyword3,
                   icq_interests_keyword4,
                   icq_moreInfo_birthDay,
                   icq_moreInfo_birthMonth,
                   icq_moreInfo_birthYear,
                   icq_moreInfo_gender,
                   icq_moreInfo_homePageAddr,
                   icq_moreInfo_lang1,
                   icq_moreInfo_lang2,
                   icq_moreInfo_lang3,
                   icq_notes,
                   icq_permissions_authRequired,
                   icq_workInfo_address,
                   icq_workInfo_city,
                   icq_workInfo_company,
                   icq_workInfo_countryCode,
                   icq_workInfo_department,
                   icq_workInfo_fax,
                   icq_workInfo_occupationCode,
                   icq_workInfo_phone,
                   icq_workInfo_position,
                   icq_workInfo_state,
                   icq_workInfo_webPage,
                   icq_workInfo_zipCode,
                   tocConfig,
                   suspendedStatus)
SELECT identScreenName,
       displayScreenName,
       authKey,
       strongMD5Pass,
       weakMD5Pass,
       confirmStatus,
       emailAddress,
       regStatus,
       isICQ,
       aim_firstName,
       aim_lastName,
       aim_middleName,
       aim_maidenName,
       aim_country,
       aim_state,
       aim_city,
       aim_nickName,
       aim_zipCode,
       aim_address,
       aim_keyword1,
       aim_keyword2,
       aim_keyword3,
       aim_keyword4,
       aim_keyword5,
       icq_affiliations_currentCode1,
       icq_affiliations_currentCode2,
       icq_affiliations_currentCode3,
       icq_affiliations_currentKeyword1,
       icq_affiliations_currentKeyword2,
       icq_affiliations_currentKeyword3,
       icq_affiliations_pastCode1,
       icq_affiliations_pastCode2,
       icq_affiliations_pastCode3,
       icq_affiliations_pastKeyword1,
       icq_affiliations_pastKeyword2,
       icq_affiliations_pastKeyword3,
       icq_basicInfo_address,
       icq_basicInfo_cellPhone,
       icq_basicInfo_city,
       icq_basicInfo_countryCode,
       icq_basicInfo_emailAddress,
       icq_basicInfo_fax,
       icq_basicInfo_firstName,
       icq_basicInfo_gmtOffset,
       icq_basicInfo_lastName,
       icq_basicInfo_nickName,
       icq_basicInfo_phone,
       icq_basicInfo_publishEmail,
       icq_basicInfo_state,
       icq_basicInfo_zipCode,
       icq_interests_code1,
       icq_interests_code2,
       icq_interests_code3,
       icq_interests_code4,
       icq_interests_keyword1,
       icq_interests_keyword2,
       icq_interests_keyword3,
       icq_interests_keyword4,
       icq_moreInfo_birthDay,
       icq_moreInfo_birthMonth,
       icq_moreInfo_birthYear,
       icq_moreInfo_gender,
       icq_moreInfo_homePageAddr,
       icq_moreInfo_lang1,
       icq_moreInfo_lang2,
       icq_moreInfo_lang3,
       icq_notes,
       icq_permissions_authRequired,
       icq_workInfo_address,
       icq_workInfo_city,
       icq_workInfo_company,
       icq_workInfo_countryCode,
       icq_workInfo_department,
       icq_workInfo_fax,
       icq_workInfo_occupationCode,
       icq_workInfo_phone,
       icq_workInfo_position,
       icq_workInfo_state,
       icq_workInfo_webPage,
       icq_workInfo_zipCode,
       tocConfig,
       suspendedStatus
FROM users_old;

DROP TABLE users_old;