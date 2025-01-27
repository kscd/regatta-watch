root := `git rev-parse --show-toplevel`

db_directory := root/".postgres"
db_port := "5433"

build-go:
    CGO_ENABLED="0" go build -o bin/main ./main

start-db:
    #!/usr/bin/env bash
    if ! `test -d {{ db_directory }}`; then
        initdb -D {{ db_directory }}
    fi
    if ! pg_ctl status -D {{ db_directory }} > /dev/null; then
        pg_ctl -D {{ db_directory }} -o "-p {{ db_port }}" -l {{ db_directory }}/logfile start
    fi

stop-db:
    #!/usr/bin/env bash
    if pg_ctl status -D {{ db_directory }} > /dev/null; then
        pg_ctl -D {{ db_directory }} stop
    fi
