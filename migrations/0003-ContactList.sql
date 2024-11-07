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
