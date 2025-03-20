-- migrate -path ./migrations -database postgres://postgres:123456@localhost/greenlight?sslmode=disable up
CREATE TABLE IF NOT EXISTS permissions(
    id bigserial PRIMARY KEY ,
    code text NOT NULL
);
CREATE TABLE IF NOT EXISTS users_permissions(
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE ,
    permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE ,
    PRIMARY KEY (user_id,permission_id)
);
-- 添加目前可能允许的两种操作-> read(1) write(2)
INSERT INTO permissions(code)
VALUES
    ('movie:read'),
    ('movie:write')