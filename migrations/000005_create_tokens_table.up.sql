-- migrate -path ./migrations -database postgres://postgres:123456@localhost/greenlight?sslmode=disable up
CREATE TABLE IF NOT EXISTS tokens(
    hash bytea PRIMARY KEY ,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE ,
    expiry timestamp(0) with time zone NOT NULL ,
    scope text NOT NULL
);