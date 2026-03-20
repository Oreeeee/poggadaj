CREATE TABLE pubdir (
    id SERIAL PRIMARY KEY,
    uin INT UNIQUE,
    firstname TEXT,
    lastname TEXT,
    nickname TEXT,
    birthyear SMALLINT,
    city TEXT,
    gender SMALLINT,
    familyname TEXT,
    familycity TEXT
);
