CREATE TABLE users (
    id serial PRIMARY KEY,
    login varchar(20) NOT NULL,
    name varchar(20) NOT NULL,
    password varchar(192) NOT NULL,
    UNIQUE(login)
)

CREATE TABLE messages (
    id serial PRIMARY KEY,
    sender_id int ON DELETE CASCADE ON UPDATE CASCADE,
    received_id int ON DELETE CASCADE ON UPDATE CASCADE,
    create_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    message text NOT NULL,
)