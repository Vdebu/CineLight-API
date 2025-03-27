-- 仅当数据库不存在时创建(数据卷postgres_data已包含旧的数据库导致CREATEDATABASEcinelight重复创建失败)
-- 安全创建数据库（兼容所有版本）
DO $$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'cinelight') THEN
            CREATE DATABASE cinelight;
        END IF;
    END $$;
-- 切换到目标数据库
\connect cinelight;
-- 建表语句
CREATE TABLE IF NOT EXISTS movies(
                                     id bigserial PRIMARY KEY ,
                                     created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
                                     title text NOT NULL ,
                                     year integer NOT NULL ,
                                     runtime integer NOT NULL ,
                                     genres text[] NOT NULL ,
                                     version integer NOT NULL DEFAULT 1
);

ALTER TABLE movies ADD CONSTRAINT movies_runtime_check CHECK ( runtime >= 0 );
ALTER TABLE movies ADD CONSTRAINT movies_year_check CHECK ( year BETWEEN 1888 AND date_part('year',NOW()));
ALTER TABLE movies ADD CONSTRAINT genres_length_check CHECK ( array_length(genres,1) BETWEEN 1 AND 5);

CREATE INDEX IF NOT EXISTS  movies_title_idx ON movies USING GIN(to_tsvector('simple',title));
CREATE INDEX IF NOT EXISTS movies_genres_idx ON movies USING GIN(genres);

CREATE TABLE IF NOT EXISTS users(
                                    id bigserial PRIMARY KEY ,
                                    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW() ,
                                    name text NOT NULL ,
                                    email citext UNIQUE NOT NULL ,
                                    password_hash bytea NOT NULL ,
                                    activated bool NOT NULL ,
                                    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS tokens(
                                     hash bytea PRIMARY KEY ,
                                     user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE ,
                                     expiry timestamp(0) with time zone NOT NULL ,
                                     scope text NOT NULL
);


CREATE TABLE IF NOT EXISTS permissions(
                                          id bigserial PRIMARY KEY ,
                                          code text NOT NULL
);
CREATE TABLE IF NOT EXISTS users_permissions(
                                                user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE ,
                                                permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE ,
                                                PRIMARY KEY (user_id,permission_id)
);

INSERT INTO permissions(code)
VALUES
    ('movie:read'),
    ('movie:write');