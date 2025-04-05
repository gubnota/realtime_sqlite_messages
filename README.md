```sh
POST /register
{
	"email": "user@example.com",
	"password": "securepassword123"
}
```

Response:

```json
{
  "email": "user@example.com",
  "id": "2c71da62-a057-4c24-beac-114b8e5d0dff"
}
{
    "error": "user already exists"
}
```

/login Response:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDM5NzQxMzMsInN1YiI6IjJjNzFkYTYyLWEwNTctNGMyNC1iZWFjLTExNGI4ZTVkMGRmZiJ9.YV0TXGsP0vCvbXZ7UAkRd1WLlAw7gVuFbdTmFpl0Kg0"
}
```

posssible errors:

```json
{ "error": "invalid credentials" }
{
    "error": "Key: 'LoginRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
{
    "error": "Key: 'LoginRequest.Password' Error:Field validation for 'Password' failed on the 'required' tag"
}
```

## Debug

```sh
go test -v ./pkg/auth
go test -v ./pkg/auth -run TestJWTMiddleware
```

```json
{
    "error": "invalid token"
}
{
  "error": "failed to send message"
}
message.go:45 database is locked
```
