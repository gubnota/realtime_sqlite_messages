import asyncio
import json
import random
import websockets
import pytest


@pytest.mark.asyncio
@pytest.mark.parametrize("n", [2, 5, 10])
async def test_send_messages(n):
    users = []
    for i in range(n):
        email = f"user{i}@example.com"
        async with websockets.connect(f"ws://localhost:8080/ws?email={email}") as ws:
            users.append((email, ws))

    for _ in range(10):
        for (email1, ws1), (email2, ws2) in random.sample(list(zip(users, users[::-1])), n):
            if ws1 is ws2:
                continue
            msg = json.dumps({"type": "message", "content": f"Hello from {email1}"})
            await ws1.send(msg)
            resp = await ws2.recv()
            assert json.loads(resp)["content"] == msg
        await asyncio.sleep(0.1)


@pytest.mark.asyncio
@pytest.mark.parametrize("n", [2, 5, 10])
async def test_subscribe_to_changes(n):
    users = []
    for i in range(n):
        email = f"user{i}@example.com"
        async with websockets.connect(f"ws://localhost:8080/ws?email={email}") as ws:
            users.append((email, ws))

    for (email1, ws1), (email2, ws2) in random.sample(list(zip(users, users[::-1])), n):
        if ws1 is ws2:
            continue
        msg = json.dumps({"type": "subscribe", "email": email2})
        await ws1.send(msg)
        resp = await ws2.recv()
        assert json.loads(resp)["type"] == "subscribe"
        assert json.loads(resp)["email"] == email1
        await asyncio.sleep(0.1)