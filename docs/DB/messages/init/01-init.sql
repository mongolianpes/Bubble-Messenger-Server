CREATE TABLE messages (
    id serial PRIMARY KEY,
    sender_id int,
    received_id int,
    create_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    message text NOT NULL
);