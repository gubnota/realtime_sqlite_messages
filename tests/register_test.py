import asyncio
import random
import string
import json
import uuid
import aiohttp
import websockets
# import websockets

BACKEND_URL = "http://localhost:8080"
WS_URL = "ws://localhost:8080/ws"

NUM_USERS = 4000  # Total number of users to register
PARALLELISM = 100  # Number of concurrent tasks

class LoadTester:
    def __init__(self):
        self.users = []
        self.connections = {}
        self.message_count = 0

    def generate_email(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=10)) + "@example.com"

    async def register_user(self, session, email, password="securepassword123"):
        async with session.post(
            f"{BACKEND_URL}/register",
            json={"email": email, "password": password}
        ) as response:
            if response.status in [200, 201, 409]:  # 201 = Created, 409 = Already exists
                print(f"Registered: {email}")
                return True
            else:
                print(f"Failed to register {email} ({response.status})")
                return False

    async def login_user(self, session, email):
        async with session.post(
            f"{BACKEND_URL}/login",
            json={"email": email, "password": "securepassword123"}
        ) as response:
            if response.status == 200:
                data = await response.json()
                return data.get("token")
        return None

    async def websocket_listener(self, email, token):
        try:
            async with websockets.connect(
                f"{WS_URL}/{email}",
                extra_headers={"Authorization": f"Bearer {token}"}
            ) as ws:
                self.connections[email] = ws
                async for message in ws:
                    self.message_count += 1
                    if self.message_count % 100 == 0:
                        print(f"Received {self.message_count} messages")
        except Exception as e:
            print(f"WebSocket error for {email}: {e}")

    async def message_sender(self):
        while True:
            await asyncio.sleep(random.uniform(0.1, 1.0))
            if self.users:
                sender = random.choice(self.users)
                receiver = random.choice([u for u in self.users if u != sender])
                message = {
                    "content": str(uuid.uuid4()),
                    "receiver": receiver["email"]
                }
                ws = sender.get("ws")
                if ws and not ws.closed:
                    await ws.send(json.dumps(message))

    async def cleanup_users(self):
        print("🧹 Cleaning up users...")
        async with aiohttp.ClientSession() as session:
            async with session.delete(f"{BACKEND_URL}/users") as resp:
                print(f"Cleanup status: {resp.status}")
                if resp.status != 200:
                    print(await resp.text())

    async def run(self):
        emails = [self.generate_email() for _ in range(NUM_USERS)]

        async with aiohttp.ClientSession() as session:
            # Parallel registration
            await asyncio.gather(*[
                self.register_user(session, email)
                for email in emails
            ])

            # Login and connect websockets
            tasks = []
            for email in emails:
                token = await self.login_user(session, email)
                if token:
                    self.users.append({"email": email, "token": token})
                    # tasks.append(self.websocket_listener(email, token))

            print(f"✅ {len(self.users)} users logged in and listening.")

            await asyncio.gather(
                self.message_sender(),
                *tasks
            )

if __name__ == "__main__":
    tester = LoadTester()
    try:
        asyncio.run(tester.run())
    except KeyboardInterrupt:
        print("🛑 Interrupted")
        asyncio.run(tester.cleanup_users())