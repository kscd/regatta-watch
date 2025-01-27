### Start the service

To build the Go service
```sh
go build
```

To run the service
```sh
go run .
```
The service is currently running on port 8091 but should later switch to the
standard port 8090.

### Ping
```sh
curl -i --location 'http://localhost:8090/ping' --header 'Content-Type: application/json'
```

/*
const pool = new Pool({
user: "regatta",
host: "localhost",
database: "regatta",
password: "1234",
port: 5432,
});
*/
