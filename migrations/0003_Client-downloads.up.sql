CREATE TABLE client_downloads (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    image_url TEXT NOT NULL,
    installer_download_url TEXT,
    extracted_download_url TEXT
);

CREATE TABLE client_downloads_descriptions (
    client_id INT NOT NULL,
    lang TEXT NOT NULL,
    description TEXT,

    PRIMARY KEY (client_id, lang),

    FOREIGN KEY (client_id) REFERENCES client_downloads(id)
        ON DELETE CASCADE
);
