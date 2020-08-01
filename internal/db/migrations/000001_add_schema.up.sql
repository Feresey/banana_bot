CREATE SCHEMA IF NOT EXISTS bot;

CREATE TABLE IF NOT EXISTS bot.persons (
    id      SERIAL  PRIMARY KEY,
    user_id INTEGER NOT NULL,
    chat_id INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS bot.warns (
    person_id INTEGER PRIMARY KEY REFERENCES bot.persons(id),
    count     INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS bot.subscriptions (
    id        SERIAL  PRIMARY KEY,
    person_id INTEGER NOT NULL REFERENCES bot.persons(id),
    chat_id   INTEGER NOT NULL
);