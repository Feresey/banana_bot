CREATE SCHEMA IF NOT EXISTS bot;

CREATE TABLE IF NOT EXISTS persons (
    id      SERIAL  PRIMARY KEY,
    user_id INTEGER NOT NULL,
    chat_id INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS warns (
    person_id INTEGER PRIMARY KEY REFERENCES persons(id),
    count     INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id        SERIAL  PRIMARY KEY,
    person_id INTEGER NOT NULL REFERENCES persons(id),
    chat_id   INTEGER NOT NULL REFERENCES persons(chat_id),
);