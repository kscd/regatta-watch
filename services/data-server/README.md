## Start the service

To build the Go service
```sh
go build
```

To run the service
```sh
go run .
```
The service is running on port 8090.

### Ping
```sh
curl -i --location 'http://localhost:8090/ping' --header 'Content-Type: application/json'
```

### Insert position
```sh
curl -i \
--location 'http://localhost:8090/pushposition' \
--header 'Content-Type: application/json' \
--data '{"positions":[{"boat": "bluebird", "longitude": 0,"latitude": 0,"measure_time": "2006-01-02T15:04:05.000Z"},
                      {"boat": "bluebird", "longitude": 1,"latitude": 1,"measure_time": "2006-01-02T15:04:06.000Z"}],
         "send_time": "2006-01-02T15:04:07.000Z"}'
```

### Extract position
```sh
curl -i \
--location 'http://localhost:8090/readposition' \
--header 'Content-Type: application/json' \
--data '{"boat": "bluebird","start_time": "2006-01-02T15:04:05.000Z","end_time": "2106-01-02T15:04:05.000Z"}'
```

### Insert Battery Level
```sh
curl -i \
--location 'http://localhost:8090/pushbattery' \
--header 'Content-Type: application/json' \
--data '"battery_level":[{"battery_level": 100,"send_time": "2006-01-02T15:04:05.000Z"},
                         {"battery_level":  90,"send_time": "2006-01-02T15:09:05.000Z"}]}'
```

### Extract Battery Level
```sh
curl -i \
--location 'http://localhost:8090/readbattery' \
--header 'Content-Type: application/json' \
--data '{"start_time": "2006-01-02T15:04:05.000Z","end_time": "2006-01-02T15:04:05.000Z"}'
```
