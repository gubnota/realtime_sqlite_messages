1. move this register users logic to a separate file / module in test:

```ini
   POST /register
   BODY: {
   "email": "user2@example.com",
   "password": "securepassword123"
   }

RPLY:
{
"email": "user3@example.com",
"id": "cab81e05-f2b1-4c82-8e51-f7595c3464df"
}
// OR
RPLY:
{
"email": "user2@example.com",
"id": "ffac3190-24d8-4e7a-8073-6bed75b4a169"
}
POST /login
BODY: {"email": "",password:""}
RPLY:
{
"expires_in": 1745231602,
"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDUyMzE2MDIsInN1YiI6ImZmYWMzMTkwLTI0ZDgtNGU3YS04MDczLTZiZWQ3NWI0YTE2OSJ9.xBJOIOQbXhC6NjIxbsQjAyKSqO47Ewjao9cKFEienwU",
"uuid": "ffac3190-24d8-4e7a-8073-6bed75b4a169"
}// OR
{
"expires_in": 1745231572,
"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDUyMzE1NzIsInN1YiI6ImNhYjgxZTA1LWYyYjEtNGM4Mi04ZTUxLWY3NTk1YzM0NjRkZiJ9.VjZU8OKnme2wHqUhUJ5HDalpAfv_AcI3aBhj30JmJcM",
"uuid": "cab81e05-f2b1-4c82-8e51-f7595c3464df"
}
```

2. move this subcribe logic to a separate file ws module in test:

2.1 subscribe all of these N\*2 users using /login results token and uuid:

Header: Authorization = Bearer eyJhbG...JmJcM
ws://localhost:8080/ws/cab81e05-f2b1-4c82-8e51-f7595c3464df

2.2 use POST {{HOST}}/send
with Auth header Bearer eyJhbG...JmJcM
and Body:

```js
{"receiver":"cab81e05-f2b1-4c82-8e51-f7595c3464df","content":"Hi"}
```

to send messages to each other
RPLY:

```js
{
"createdAt": 1745145917,
"delivered": false,
"id": 1
}
```

2.4 on /ws/cab81e05-f2b1-4c82-8e51-f7595c3464df you should get a message:

```js
{"data":{"content":"Hi","createdAt":1745145997,"id":2,"sender":"ffac3190-24d8-4e7a-8073-6bed75b4a169"},"type":"message"}
```

2.3 use GET {{HOST}}/messages?from=1745145997 with Authorization header Bearer eyJhbG...JmJcM for fetching received messages to compare with those from /ws/
for sender it will be empty:

```js
{
    "messages": []
}
```

for receiver:

```js
{
    "messages": [
        {
            "content": "Hi",
            "createdAt": 1745145997,
            "delivered": true,
            "id": 2,
            "sender": "ffac3190-24d8-4e7a-8073-6bed75b4a169"
        }
    ]
}
```

write tests/ws_test.py

1. it has params N pairs of users
2. it registers N \* 2 users and stores to the Map<String,List<String>> all of these users uuids, emails, passwords (is the same password for each of these user)
3. it subsribes to /ws/{uuid} of each user
4. it tests that messages are received

Thus I need a test with n pairs of users param and no of sending message per minute within each pair of users.
a message sent from a user to b should be received on b side.
