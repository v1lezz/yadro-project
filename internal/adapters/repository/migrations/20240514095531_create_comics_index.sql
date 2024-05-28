-- +goose Up
-- +goose StatementBegin
CREATE TABLE comics (
    id INTEGER PRIMARY KEY,
    image_url VARCHAR(255) NOT NULL
);

CREATE TABLE keyword (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    UNIQUE(keyword)
);

CREATE TABLE comics_keyword (
    id SERIAL PRIMARY KEY,
    comics_id INTEGER REFERENCES comics(id) ON UPDATE RESTRICT ON DELETE CASCADE,
    keyword_id SERIAL REFERENCES keyword(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    UNIQUE(comics_id, keyword_id)
);

CREATE TABLE index (
    id SERIAL PRIMARY KEY,
    keyword_id SERIAL NOT NULL REFERENCES keyword(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    comics_id INTEGER NOT NULL REFERENCES keyword(id) ON UPDATE RESTRICT ON DELETE CASCADE,
    UNIQUE(keyword_id, comics_id)
);

CREATE TABLE time (
    id SERIAL PRIMARY KEY,
    update_time_comics TIMESTAMP NOT NULL,
    update_time_index TIMESTAMP NOT NULL,
    last_full_check_time TIMESTAMP NOT NULL
);

INSERT INTO time(update_time_comics, update_time_index, last_full_check_time) VALUES
('2000-01-01 00:00:00','2000-01-01 00:00:00','2000-01-01 00:00:00');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE index;
DROP TABLE comics_keyword;
DROP TABLE keyword;
DROP TABLE comics;
DROP TABLE time;
-- +goose StatementEnd
