-- Schema
CREATE TABLE gguser (
    uin SERIAL PRIMARY KEY,
    password_gg32 BIGINT,
    password_sha1 VARCHAR(20),
    notify_list INTEGER[]
);

CREATE TABLE adserver_ad (
    id SERIAL PRIMARY KEY,
    adtype SMALLINT,
    bannertype SMALLINT,
    image TEXT,
    html TEXT
);


-- Initial data
INSERT INTO gguser
    (uin, password_gg32, password_sha1)
    VALUES
    (1, 4105424095, NULL); -- Creates an initial user with the number 1 and password 123

INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 0, 'Hello from poggadaj-HTTP!');

INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 0, 'This is one of the banner responses.');

INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 0, 'Lorem ipsum');

INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 1, 'Hello from poggadaj-HTTP!');

INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 1, 'This is one of the small banner responses.');

INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 1, 'Lorem ipsum');

INSERT INTO adserver_ad
    (adtype, bannertype, html)
    VALUES
    (0, 2, 'Hello from poggadaj-HTTP!');

INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 2, 'This is one of the main banner responses.');

INSERT INTO adserver_ad
(adtype, bannertype, html)
VALUES
    (0, 2, 'Lorem ipsum');
