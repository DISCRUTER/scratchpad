CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL CONSTRAINT users_uc_email UNIQUE,
    hashed_password CHAR(60) NOT NULL,
    created TIMESTAMPTZ NOT NULL
);
