CREATE TABLE IF NOT EXISTS "position" (
    id BIGSERIAL PRIMARY KEY,
    longitude pg_catalog.float8 NOT NULL DEFAULT FALSE,
    latitude pg_catalog.float8 NOT NULL DEFAULT FALSE,
    measure_time timestamptz NOT NULL DEFAULT FALSE,
    send_time timestamptz NOT NULL DEFAULT FALSE,
    receive_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "battery" (
    id BIGSERIAL PRIMARY KEY,
    level pg_catalog.float8 NOT NULL DEFAULT FALSE,
    send_time timestamptz NOT NULL DEFAULT FALSE,
    receive_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
