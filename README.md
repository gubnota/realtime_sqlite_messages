## Generate certs

```sh
openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out cert.pem -days 365 -subj "/CN=localhost"
```

## Run tests

```sh
go test -v -run TestFileChangeNotification
```
