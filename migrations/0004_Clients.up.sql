CREATE TABLE clients (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    image_url TEXT,
    priority INT NOT NULL
);

CREATE TABLE client_downloads (
    id SERIAL PRIMARY KEY,
    client_id INT NOT NULL,
    file_variant SMALLINT NOT NULL,
    url TEXT NOT NULL,

    CONSTRAINT fk_client_downloads_client
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE
);

CREATE TABLE client_descriptions (
    id SERIAL PRIMARY KEY,
    client_id INT NOT NULL,
    language VARCHAR(2) NOT NULL,
    description TEXT NOT NULL,

    CONSTRAINT fk_client_descriptions_client
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE
);
