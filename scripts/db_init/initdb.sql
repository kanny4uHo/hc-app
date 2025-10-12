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

------------------------------------------------------------

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

------------------------------------------------------------

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

------------------------------------------------------------

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

------------------------------------------------------------

CREATE DATABASE inventoryapp;
\c inventoryapp;

CREATE TABLE IF NOT EXISTS items
(
    id          text primary key,
    description text not null,
    amount      int  not null default 0 check ( amount > 0 )
);

CREATE TABLE IF NOT EXISTS reservations
(
    id       serial PRIMARY KEY,
    order_id int  not null,
    item_id  text not null,
    amount   int  not null
);

CREATE USER inventoryapp WITH PASSWORD 'laksjdlkdsjafhlkasdjlhasd';
GRANT ALL ON TABLE items TO inventoryapp;
GRANT ALL ON TABLE reservations TO inventoryapp;
GRANT USAGE, SELECT ON SEQUENCE reservations_id_seq TO inventoryapp;

-- тестовые данные
INSERT INTO items(id, description, amount) VALUES ('pencil', 'Red pencil, make of wood', 10);

------------------------------------------------------------

CREATE DATABASE deliveryapp;
\c deliveryapp;

CREATE TABLE IF NOT EXISTS couriers
(
    id          serial PRIMARY KEY,
    name        text not null,
    is_on_shift int default 0
);

CREATE INDEX couriers_is_on_shift_idx ON couriers (is_on_shift);

CREATE TABLE IF NOT EXISTS deliveries
(
    id         serial PRIMARY KEY,
    order_id   int  not null,
    courier_id int  not null,
    status     text not null
);

CREATE INDEX deliveries_courier_id_idx ON deliveries (courier_id);
CREATE INDEX deliveries_status_idx ON deliveries (status);

CREATE USER deliveryapp WITH PASSWORD 'qpwoeinmnzkjasdlaksdjmxwp';

GRANT ALL ON TABLE couriers TO deliveryapp;
GRANT USAGE, SELECT ON SEQUENCE couriers_id_seq TO deliveryapp;

GRANT ALL ON TABLE deliveries TO deliveryapp;
GRANT USAGE, SELECT ON SEQUENCE deliveries_id_seq TO deliveryapp;

-- тестовые данные
INSERT INTO couriers (name, is_on_shift) VALUES ('Курьер1', 1);
INSERT INTO couriers (name, is_on_shift) VALUES ('Курьер2', 0);
INSERT INTO couriers (name, is_on_shift) VALUES ('Курьер3', 1);
INSERT INTO couriers (name, is_on_shift) VALUES ('Курьер4', 1);
INSERT INTO couriers (name, is_on_shift) VALUES ('Курьер5', 0);
INSERT INTO couriers (name, is_on_shift) VALUES ('Курьер6', 1);




