## Run book

First start the backends with the command
```sh
go run .
```
Then start the frontend of the website
```sh
pnpm run dev
```

The backend has the 



## Make new database ready (mac)

Initialise postgres
```sh
initdb -D Documents/regatta-watch/postgres
```

Start new database service (starts in the background)
```sh
pg_ctl -D Documents/regatta-watch/postgres -l logfile start
```

Create new postgres database
```sh
createdb regatta
```

The program `psql` can be used to execute SQL commands in an CLI. To start
`psql`, you need a postgres account with the same name as your computer
username. 'Log in' to the database by running:
```sh
psql regatta
```

Inside the postgres CLI create a new table for the data server
```postgresql
CREATE TABLE IF NOT EXISTS positions_data_server (
    id BIGSERIAL PRIMARY KEY,
    boat text NOT NULL DEFAULT '',
    longitude pg_catalog.float8 NOT NULL DEFAULT 0.0,
    latitude pg_catalog.float8 NOT NULL DEFAULT 0.0,
    measure_time timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
    send_time timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
    receive_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```
And do the same for the website backend which serves as a local copy to reduce
the required bandwidth
```postgresql
CREATE TABLE IF NOT EXISTS positions_website_backend (
    id BIGSERIAL PRIMARY KEY,
    boat text NOT NULL DEFAULT '',
    longitude pg_catalog.float8 NOT NULL DEFAULT 0.0,
    latitude pg_catalog.float8 NOT NULL DEFAULT 0.0,
    measure_time timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
    send_time timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
    receive_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

List the tables
```postgresql
\dt
```

Create a role for the services to log in with
```postgresql
CREATE ROLE regatta WITH LOGIN PASSWORD '1234';
GRANT CONNECT ON DATABASE regatta TO regatta;
GRANT CREATE ON DATABASE regatta TO regatta;
GRANT USAGE ON SCHEMA public TO regatta;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO regatta;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO regatta;
```
and give it the rights it needs in the database.

## Other useful SQL commands

Take a look into the data server database
```postgresql
SELECT * FROM positions_data_server;
```