-- Schema
CREATE TABLE gguser (
    uin SERIAL PRIMARY KEY,
    password_gg32 BIGINT,
    password_sha1 VARCHAR(20),
    notify_list INTEGER[]
);

-- Initial data

-- Creates an initial user with the number 1 and password 123
INSERT INTO gguser
    (uin, password_gg32, password_sha1)
    VALUES
    (1, 4105424095, NULL);
