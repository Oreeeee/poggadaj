CREATE TABLE gguser (
    uin SERIAL PRIMARY KEY,
    password_gg32 BIGINT,
    password_sha1 VARCHAR(20)
);
