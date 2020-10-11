# LoRaWAN ID Generator

## Potential improvements
Persistent storage cache for registered ID's so we don't have to keep hitting the API.

Exit clause for Generator.Generate when checking short-form uniqueness. Could result in 
an endless loop if seeder doesn't have enough entropy.

Tests could be much more thorough but this is an assessment, not production code.

API could be improved with CMD flags to set bind address. Defaults to localhost:8080.

## Building CLI
```
go build -o cli ./cmd/cli/cli.go
```

## Running CLI
```
./cli > ids.txt
cat ids.txt
```

## Building API Server
```
go build -o api ./cmd/api/api.go
```

## Accessing API server
```
{idempotencyKey} = 8-16 character long HEX value
curl http:localhost//:8080/ids/{idempotencyKey}
```

## Testing
```
go test -count=1 -v ./...
```