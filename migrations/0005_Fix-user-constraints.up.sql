ALTER TABLE gguser DROP CONSTRAINT gguser_name_key;
ALTER TABLE gguser DROP CONSTRAINT gguser_email_key;

CREATE UNIQUE INDEX gguser_name_unique ON gguser (lower(name));
CREATE UNIQUE INDEX gguser_email_unique ON gguser (lower(email));
