CREATE TABLE users (
    id serial PRIMARY KEY,
    login varchar(20) NOT NULL,
    name varchar(20) NOT NULL,
    password varchar(192) NOT NULL,
    avatar_path varchar(64),
    UNIQUE(login)
);