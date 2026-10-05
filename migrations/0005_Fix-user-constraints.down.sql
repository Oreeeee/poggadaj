DROP INDEX gguser_name_unique;
DROP INDEX gguser_email_unique;

ALTER TABLE gguser ADD CONSTRAINT gguser_name_key UNIQUE (name);
ALTER TABLE gguser ADD CONSTRAINT gguser_email_key UNIQUE (email);
