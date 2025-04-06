## Simple backend for local testing

- uses SQLite as a database
- utilizes websockets for real-time messaging
- uses JWT for authentication

### Installation

Create `.env`:

```ini
SQLITE_PATH=sq.db
JWT_SECRET=TlFuT3JUMWNXano4N2pVN0FmU3BuamRUdFNTTzAzMndBQzRmN1BBemtlbz0K
PORT=8080
```

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
```

Or for error:

```json
{
  "error": "user already exists"
}
```

`POST /login` Response:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDM5NzQxMzMsInN1YiI6IjJjNzFkYTYyLWEwNTctNGMyNC1iZWFjLTExNGI4ZTVkMGRmZiJ9.YV0TXGsP0vCvbXZ7UAkRd1WLlAw7gVuFbdTmFpl0Kg0"
}
```

posssible errors:

```json
{ "error": "invalid credentials" }
```

```json
{
  "error": "Key: 'LoginRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
```

```json
{
  "error": "Key: 'LoginRequest.Password' Error:Field validation for 'Password' failed on the 'required' tag"
}
```

## Debug and test

```sh
go test -v ./pkg/auth
go test -v ./pkg/auth -run TestJWTMiddleware
./run_load_test.sh
```

## Possible errors

```json
{
  "error": "invalid token"
}
```

If database file is locked by another process:

```json
{
  "error": "failed to send message"
}
```

## Generate token

Possible reason for invalid token error is incorrect secret:

```sh
openssl rand -base64 32 | base64 -w 0
TlFuT3JUMWNXano4N2pVN0FmU3BuamRUdFNTTzAzMndBQzRmN1BBemtlbz0K
```

Solved the problem.

```json
{
  "data": {
    "content": "Hi",
    "createdAt": "2025-04-06T17:40:32.15988+03:00",
    "id": 1,
    "sender": "2c71da62-a057-4c24-beac-114b8e5d0dff"
  },
  "type": "message"
}
```

```sh
# Example request
# GET /messages?from=2024-01-01T00:00:00Z
GET /messages?from=1672531200  # 2023-01-01 00:00:00 UTC
Authorization: Bearer <your_token>
```

```json
{
  "messages": [
    {
      "id": 123,
      "sender": "user-uuid",
      "content": "Hello",
      "createdAt": 1672531265,
      "delivered": true
    }
  ]
}
```
