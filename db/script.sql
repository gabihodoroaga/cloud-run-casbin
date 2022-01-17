CREATE DATABASE IF NOT EXISTS cloudrun-casbin;

CREATE TABLE IF NOT EXISTS users_roles (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL,
    role VARCHAR(100) NOT NULL,
    CONSTRAINT ck_unique_role UNIQUE(email,role)
);

INSERT INTO users_roles (email, role) VALUES ('admin@example.com', 'admin');
