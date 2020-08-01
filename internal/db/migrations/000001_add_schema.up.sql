CREATE SCHEMA IF NOT EXISTS bot;

CREATE TABLE IF NOT EXISTS bot.persons (
    person_id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS bot.warns (
    warn_person_id BIGINT PRIMARY KEY REFERENCES bot.persons(person_id),
    warn_count     BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS bot.subscriptions (
    sub_id BIGINT PRIMARY KEY REFERENCES bot.persons(person_id)
);