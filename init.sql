-- Schema
CREATE SEQUENCE uin_seq
MINVALUE 1
START 101
INCREMENT 1
AS INT;

CREATE TABLE gguser (
    id SERIAL PRIMARY KEY,
    uin INT UNIQUE,
    password_gg_ancient BIGINT,
    password_gg32 BIGINT,
    password_sha1 VARCHAR(40),
    name TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    joined TIMESTAMP DEFAULT now()
);

CREATE TABLE adserver_ad (
    id SERIAL PRIMARY KEY,
    adtype SMALLINT,
    bannertype SMALLINT,
    image TEXT,
    html TEXT
);

CREATE TABLE ggcontact (
    id SERIAL PRIMARY KEY,
    owner_uin INT NOT NULL,
    firstname TEXT,
    lastname TEXT,
    pseudonym TEXT,
    display_name TEXT,
    mobile_number BIGINT,
    grp TEXT,
    uin INT,
    email TEXT,
    avail_sound SMALLINT,
    avail_path TEXT,
    msg_sound SMALLINT,
    msg_path TEXT,
    hidden BOOLEAN,
    landline_number BIGINT,
    
    -- Make sure not to store duplicate contacts
    UNIQUE (
        owner_uin,
        firstname,
        lastname,
        pseudonym,
        display_name,
        mobile_number,
        grp,
        uin,
        email,
        avail_sound,
        avail_path,
        msg_sound,
        msg_path,
        hidden,
        landline_number
    )
);

-- Initial data
INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 0, 'Hello from poggadaj-HTTP!');

INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 1, 'Hello from poggadaj-HTTP!');

INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 2, 'Hello from poggadaj-HTTP!');
