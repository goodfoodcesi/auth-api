CREATE EXTENSION IF NOT EXISTS  "uuid-ossp";

CREATE TYPE USER_ROLE AS ENUM ('admin', 'user', 'manager', 'driver');

CREATE TABLE users (
    "user_id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(500) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    user_role user_role NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id)
);
-- email is unique
CREATE UNIQUE INDEX idx_users_email ON users(email);
-- phone_number is unique
CREATE UNIQUE INDEX idx_users_phone_number ON users(phone_number);
-- user_role index
CREATE INDEX idx_users_user_role ON users(user_role);




