import asyncio
import random
import json
import uuid
import websockets
import aiohttp

BACKEND_URL = "http://localhost:8080"
WS_URL = "ws://localhost:8080/ws"

class LoadTester:
    def __init__(self):
        self.users = []
        self.connections = {}
        self.message_count = 0

    async def login_user(self, session, email):
        async with session.post(
            f"{BACKEND_URL}/login",
            json={"email": email, "password": "testpassword123"}
        ) as response:
            if response.status == 200:
                data = await response.json()
                return data.get("token")
        return None

    async def websocket_listener(self, email, token):
        async with websockets.connect(
            f"{WS_URL}/{email}",
            headers={"Authorization": f"Bearer {token}"}
        ) as ws:
            self.connections[email] = ws
            async for message in ws:
                self.message_count += 1
                if self.message_count % 100 == 0:
                    print(f"Received {self.message_count} messages")

    async def message_sender(self):
        while True:
            await asyncio.sleep(random.uniform(0.1, 1.0))  # Adjust message rate
            if self.users:
                sender = random.choice(self.users)
                receiver = random.choice([u for u in self.users if u != sender])
                
                message = {
                    "content": str(uuid.uuid4()),
                    "receiver": receiver["email"]
                }
                
                if sender["ws"] and not sender["ws"].closed:
                    await sender["ws"].send(json.dumps(message))

    async def run(self):
        async with aiohttp.ClientSession() as session:
            # Load test users
            with open("test_users.txt") as f:
                emails = [line.strip() for line in f.readlines()]
            
            # Login and establish WS connections
            tasks = []
            for email in emails:
                token = await self.login_user(session, email)
                if token:
                    self.users.append({"email": email, "token": token})
                    tasks.append(self.websocket_listener(email, token))
            
            # Start sender and wait
            await asyncio.gather(
                self.message_sender(),
                *tasks
            )

if __name__ == "__main__":
    tester = LoadTester()
    asyncio.run(tester.run())