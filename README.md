## Simple backend for self-hosted chat app with notifications

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

If you're running w/o Docker:

```sh
go mod download
go install
go run main.go
```

- run tests:

```sh
uv run tests/ws_test.py
```

### User module

```sh
=== 1. Register with invalid data ===
400 {'error': "Key: 'Email' Error:Field validation for 'Email' failed on the 'email' tag\nKey: 'Password' Error:Field validation for 'Password' failed on the 'min' tag"}

=== 2. Login with incorrect credentials ===
401 {"error":"invalid credentials"}

Register sender: 201 {'email': 'd7334c70@example.com', 'id': '7bdfe0f6-b4bd-4e67-8079-ed0d8907698c'}
Register receiver: 201 {'email': '04a4f907@example.com', 'id': 'a8c09f46-39c1-4fe5-9e99-464e1f831caa'}

=== 3. Login with correct credentials ===
200 {'expires_in': 1746453282, 'token': 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY0NTMyODIsInN1YiI6IjdiZGZlMGY2LWI0YmQtNGU2Ny04MDc5LWVkMGQ4OTA3Njk4YyJ9.V_y76Le7TOrHIwVkab61oqCv0ZAkPiKolVDO05PcLFA', 'uuid': '7bdfe0f6-b4bd-4e67-8079-ed0d8907698c'}

=== 4. Send message with incorrect token ===
401 {"error":"Invalid or expired token"}

=== 5. Send message with correct token ===
400 {'error': "Key: 'MessageRequest.Receiver' Error:Field validation for 'Receiver' failed on the 'uuid4' tag"}

=== 6. Forgot password with incorrect email ===
200 {'status': 'reset link sent if email exists'}

=== 7. Forgot password with correct email ===
200 {'status': 'reset link sent if email exists'}

=== 8. Reset password with incorrect token ===
401 {"error":"invalid or expired token"}

=== 9. Reset password with correct token ===
200 {'status': 'password reset successful'}
```

touch `sq.db` (otherwise it will create a directory)

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

## Debugging

- [localhost:6060/debug/pprof/](localhost:6060/debug/pprof/)

## TODO

- [ ] callback to send an email via python microservice that runs on another port to reset password (can be used for password recovery only 10 minutes). POST /reset-password {email: string} (email webhook put to .env)
- [ ] same callback for external push notifications service (push_webhook) (these all put to .env)
- [ ] add to devices table active devices (last_seen, change status to offline when device is not active i.e. /ws/ is closed)
- [ ] add last seen field to users table
- [ ] when a message is delivered (=1), all prev messages from this users are delivered
- [ ] change created_at everywhere to timestamp(seconds)
- [ ] add index to receiver, sender
- [ ] add table score for users (int) or int score column for users table

// push_webhook: Main push method sender
Bun.serve({
port: 8000,
hostname: "0.0.0.0", // listen at any address
fetch: async (req) => {
try {
if (req.method == "POST") {
message = {
title: `${senderName} invites you`,
body: "You've received an invitation. Do you accept or refuse?",
};

## update score

```sh
POST /update-score
Authorization: Bearer <token>
Content-Type: application/json

{
	"score": 150
}
```

## Device id added for push notifications

```sh
X-Device-ID: <headerid>

# Connect with device header
curl -H "X-Device-ID: mobile-123" -H "Authorization: Bearer ..." http://localhost:8080/ws/user-uuid
```

## How to implement voting game the easy way?

- {vote id}, receiver | sender in (messages /ws (type "game" instead of "message"))
- after game is created there is countdown 2h, all points of reward will get the voted user
- if one user upvotes, the other downvotes, first gets 0, second gets 5
- if one user upvotes, the other upvotes, first gets 3, second gets 3
- if one user downvotes, the other downvotes, first gets 1, second gets 1
- when the 2nd user votes as well, 1st user receives a /ws message: {game id} that means game is over

### game table

| name     | type   | description    | default |
| -------- | ------ | -------------- | ------- |
| id       | bigint | index          | auto    |
| sender   | uuid   | game initiator |         |
| receiver | uuid   | game receiver  |         |
| created  | int    | timestamp      |         |
| svote    | int    | sender_vote    | 0       |
| rvote    | int    | receiver_vote  | 0       |
| status   | string | open, closed   |         |

- sender_vote|receiver_vote is -1 or 1 after voted (0 initially)
- don't forget to update users scores accroding to above written rules

### REST methods

- POST /game/invite

```json
{
  "receiver": "550e8400-e29b-41d4-a716-446655440000"
}
```

on /ws/:

```json
{
  "game": {
    "id": 1,
    "sender": "bafe9ff8-6ee4-49b1-8d56-7ff677c50d1b",
    "receiver": "ddd9bf62-9b47-4d7e-997e-624f21c21964",
    "created": "2025-04-08T23:01:46.198444+03:00"
  },
  "type": "game_invite"
}
```

- POST /game/vote

```json
{
  "game_id": 123,
  "vote": 1
}
```

```json
//sender vote
{
  "ID": 1,
  "Sender": "bafe9ff8-6ee4-49b1-8d56-7ff677c50d1b",
  "Receiver": "ddd9bf62-9b47-4d7e-997e-624f21c21964",
  "Created": "2025-04-08T23:01:46.198444+03:00",
  "Svote": -1,
  "Rvote": 0,
  "Status": "open"
}

//receiver vote
{
    "ID": 1,
    "Sender": "bafe9ff8-6ee4-49b1-8d56-7ff677c50d1b",
    "Receiver": "ddd9bf62-9b47-4d7e-997e-624f21c21964",
    "Created": "2025-04-08T23:01:46.198444+03:00",
    "Svote": -1,
    "Rvote": 1,
    "Status": "closed"
}
```

```json
{
  "error": "vote failed"
}
{
    "error": "invalid vote operation"
}
```

- /ws/ will receive:

```json
// after game creation
{
    "game": {
        "id": 2,
        "sender": "ddd9bf62-9b47-4d7e-997e-624f21c21964",
        "receiver": "bafe9ff8-6ee4-49b1-8d56-7ff677c50d1b",
        "created": "2025-04-08T23:23:34.971084+03:00"
    },
    "type": "game_invite"
}

{
  "type": "game_result",
  "game": {
    "id": 123,
    "status": "closed",
    "svote": 1,
    "rvote": -1
  }
}
// after recieving vote (finished):
{"game":{"ID":1,"Sender":"bafe9ff8-6ee4-49b1-8d56-7ff677c50d1b","Receiver":"ddd9bf62-9b47-4d7e-997e-624f21c21964","Created":"2025-04-08T23:01:46.198444+03:00","Svote":-1,"Rvote":1,"Status":"closed"},"type":"game_result"}

```

### GET /games/active (Auth)

```json
[
  {
    "id": 1,
    "sender": "bafe9ff8-6ee4-49b1-8d56-7ff677c50d1b",
    "receiver": "ddd9bf62-9b47-4d7e-997e-624f21c21964",
    "created_at": "2025-04-08T23:01:46.198444+03:00",
    "status": "open"
  }
]
```

### GET /result (only finished games)

```json
[]
```

## Local run and test

1. make sure `.env` file exists and contains all vars.

```sh
export GIN_MODE=release; ./realtime_sqlite_messages
```

## Memory consumption

1. AutoMigrate only runs if the environment variable RUN_MIGRATIONS=1 is set.
2. The periodic cleanup task is now stoppable and won’t leak goroutines.
3. A graceful shutdown flow was added:

```shell
initially, it's 10 threads, 10.0 MB
```

## Known issues

TODO:
when /send a message make sure user exists only then write it to DB

```js
{
"createdAt": 1745145917,
"delivered": false,
"id": 1
}
```

## Limitations

In order to handle more simultaneous connections make sure to increase ulimit

    - Use 127.0.0.1 instead of localhost to eliminate DNS resolution failures under heavy load.
    - Limit concurrent WebSocket connects with a semaphore (MAX_CONCURRENT).
    - Implement retry logic for each WS connection.
    - Provide tuning guidance in the README (e.g., ulimit -n).
