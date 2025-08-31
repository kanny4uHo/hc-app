CREATE DATABASE userapp;
\c userapp;
CREATE TABLE IF NOT EXISTS users
(
    id            serial PRIMARY KEY,
    username      TEXT not null,
    password_hash TEXT not null,
    email         TEXT not null,
    first_name    TEXT not null,
    last_name     TEXT not null
);

CREATE UNIQUE INDEX users_username_idx ON users (username);
CREATE UNIQUE INDEX users_email_idx ON users (email);

CREATE USER userapp WITH PASSWORD 'asdajlshdfljasdlkjasbdf';
GRANT ALL ON TABLE users TO userapp;
GRANT USAGE, SELECT ON SEQUENCE users_id_seq TO userapp;


CREATE DATABASE orderapp;
\c orderapp;

CREATE TABLE IF NOT EXISTS orders
(
    id      serial PRIMARY KEY,
    item    TEXT not null,
    price   INT  not null,
    status  TEXT not null,
    user_id INT  not null
);

CREATE USER orderapp WITH PASSWORD 'qocnhfoiucjmwfoijwamxaos';
GRANT ALL ON TABLE orders TO orderapp;
GRANT USAGE, SELECT ON SEQUENCE orders_id_seq TO orderapp;


CREATE DATABASE notificationapp;
\c notificationapp;
CREATE TABLE IF NOT EXISTS notifications
(
    id              serial PRIMARY KEY,
    timestamp       timestamp not null,
    order_id        INT       NOT NULL,
    recipient_email TEXT      not null,
    message         TEXT      not null
);

CREATE USER notificationapp WITH PASSWORD 'cqpmoiapvjmpaoinwejcanpwocfj';
GRANT ALL ON TABLE notifications TO notificationapp;
GRANT USAGE, SELECT ON SEQUENCE notifications_id_seq TO notificationapp;

CREATE DATABASE billingapp;
\c billingapp;

CREATE TABLE IF NOT EXISTS accounts
(
    id      serial PRIMARY KEY,
    user_id INT NOT NULL,
    balance INT NOT NULL default 0
);

CREATE USER billingapp WITH PASSWORD 'adsklkfnasdjkfnwkjneqwnje';
GRANT ALL ON TABLE accounts TO billingapp;
GRANT USAGE, SELECT ON SEQUENCE accounts_id_seq TO billingapp;




