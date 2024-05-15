-- +goose Up
-- +goose StatementBegin
CREATE TABLE comics (
    id INTEGER PRIMARY KEY,
    image_url VARCHAR(255) NOT NULL,
);

CREATE TABLE keyword (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL
);

CREATE TABLE comics_keyword (
    id SERIAL PRIMARY KEY,
    comics_id INTEGER REFERENCES comics(id) ON UPDATE RESTRICT ON DELETE CASCADE,
    keyword_id SERIAL REFERENCES keyword(id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE TABLE index (
    id SERIAL PRIMARY KEY,
    keyword_id SERIAL NOT NULL REFERENCES keyword(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    comics_id INTEGER NOT NULL REFERENCES keyword(id) ON UPDATE RESTRICT ON DELETE CASCADE
);

CREATE TABLE time (
    id SERIAL PRIMARY KEY,
    update_time_comics TIME NOT NULL,
    update_time_index TIME NOT NULL,
    last_full_check_time TIME NOT NULL
);

INSERT INTO time(update_time, update_time_index, last_full_check_time) VALUES
(0,0,0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE index;
DROP TABLE comics_keyword;
DROP TABLE keyword;
DROP TABLE comics;
DROP TABLE time;
-- +goose StatementEnd