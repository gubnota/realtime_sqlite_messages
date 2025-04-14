
import asyncio
import random
import string
import json
import uuid
import aiohttp
import websockets

BACKEND_URL = "http://localhost:8080"
WS_URL = "ws://localhost:8080/ws"
NUM_USERS = 100
DELAY = 0.1  # in seconds

class LoadTester:
    def __init__(self):
        self.users = []  # [{email, uuid, token, ws}]
        self.message_count = 0

    def generate_email(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=10)) + "@example.com"

    async def register_user(self, session, email, password="test123456"):
        async with session.post(
            f"{BACKEND_URL}/register",
            json={"email": email, "password": password}
        ) as response:
            if response.status in [200, 201, 409]:
                print(f"✅ Registered {email}")
                return True
            else:
                print(f"❌ Failed to register {email} ({response.status})")
                return False

    async def login_user(self, session, email):
        async with session.post(
            f"{BACKEND_URL}/login",
            json={"email": email, "password": "test123456"}
        ) as response:
            if response.status == 200:
                data = await response.json()
                return data.get("token"), data.get("uuid")
        return None, None

    async def connect_ws(self, token, uuid_):
        headers = {
            "Authorization": f"Bearer {token}",
            "X-Device-ID": uuid_
        }
        try:
            ws = await websockets.connect(f"{WS_URL}/{uuid_}", additional_headers=headers)
            return ws
        except Exception as e:
            print(f"⚠️ WebSocket connection failed for {uuid_}: {e}")
            return None

    async def listen_ws(self, user):
        ws = user["ws"]
        try:
            async for message in ws:
                self.message_count += 1
                print(f"📨 {user['uuid']} received: {message}")
        except Exception as e:
            print(f"WebSocket closed for {user['uuid']}: {e}")

    async def send_messages(self):
        while True:
            await asyncio.sleep(DELAY)
            if len(self.users) < 2:
                continue
            u1, u2 = random.sample(self.users, 2)
            payload = {
                "receiver": u2["uuid"],
                "content": ''.join(random.choices(string.ascii_letters, k=16))
            }
            headers = {"Authorization": f"Bearer {u1['token']}"}
            async with aiohttp.ClientSession() as session:
                async with session.post(f"{BACKEND_URL}/send", json=payload, headers=headers) as resp:
                    if resp.status != 201:
                        print(f"❌ Failed to send from {u1['uuid']} to {u2['uuid']}")

    async def cleanup_users(self):
        async with aiohttp.ClientSession() as session:
            async with session.delete(f"{BACKEND_URL}/users") as resp:
                print(f"🧹 Cleanup status: {resp.status}")

    async def run(self):
        emails = [self.generate_email() for _ in range(NUM_USERS)]
        async with aiohttp.ClientSession() as session:
            await asyncio.gather(*[self.register_user(session, email) for email in emails])

            for email in emails:
                token, uuid_ = await self.login_user(session, email)
                if token and uuid_:
                    ws = await self.connect_ws(token, uuid_)
                    if ws:
                        user = {"email": email, "token": token, "uuid": uuid_, "ws": ws}
                        self.users.append(user)
                        asyncio.create_task(self.listen_ws(user))

        print(f"🔗 {len(self.users)} users connected. Sending messages...")
        await self.send_messages()

if __name__ == "__main__":
    tester = LoadTester()
    try:
        asyncio.run(tester.run())
    except KeyboardInterrupt:
        print("🛑 Interrupted, cleaning up...")
        asyncio.run(tester.cleanup_users())
