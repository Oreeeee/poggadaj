CREATE TABLE clients (
    id INT PRIMARY KEY,
    name TEXT,
    image_url TEXT
);

CREATE TABLE client_downloads (
    id SERIAL PRIMARY KEY,
    client_id INT,
    file_variant SMALLINT,
    url TEXT,

    CONSTRAINT fk_client_downloads_client
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE
);

CREATE TABLE client_descriptions (
    id SERIAL PRIMARY KEY,
    client_id INT,
    language VARCHAR(2),
    description TEXT,

    CONSTRAINT fk_client_descriptions_client
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE
);
