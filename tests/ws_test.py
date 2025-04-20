import asyncio
import random
import string
import aiohttp
import websockets
import resource

# Increase file descriptor limit (soft) to handle many connections
soft, hard = resource.getrlimit(resource.RLIMIT_NOFILE)
new_soft = min(10000, hard)
resource.setrlimit(resource.RLIMIT_NOFILE, (new_soft, hard))
print(f"Adjusted RLIMIT_NOFILE: soft={new_soft}, hard={hard}")

WS_HOST = "127.0.0.1"
WS_PORT = 8080

BACKEND_URL = f"http://{WS_HOST}:{WS_PORT}"
WS_URL = f"ws://{WS_HOST}:{WS_PORT}/ws"
NUM_USERS = 200
DELAY = 0.1  # in seconds
MAX_CONCURRENT = 20  # lower concurrency to reduce spikes
CONNECT_RETRIES = 3
RETRY_DELAY = 0.1

class LoadTester:
    def __init__(self):
        self.users = []  # [{email, uuid, token, ws}]
        self.message_count = 0
        self.sem = asyncio.Semaphore(MAX_CONCURRENT)

    def generate_email(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=10)) + "@example.com"

    async def register_user(self, session, email, password="test123456"):
        async with session.post(
            f"{BACKEND_URL}/register",
            json={"email": email, "password": password}
        ) as response:
            if response.status in [200, 201, 409]:
                print(f"✅ Registered {email}")
            else:
                print(f"❌ Failed to register {email} ({response.status})")

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
        url = f"{WS_URL}/{uuid_}"
        headers = {
            "Authorization": f"Bearer {token}",
            "X-Device-ID": uuid_
        }
        last_exc = None
        for attempt in range(1, CONNECT_RETRIES + 1):
            try:
                ws = await websockets.connect(url, additional_headers=headers)
                print(f"🔗 WS connected for {uuid_}")
                return ws
            except Exception as e:
                last_exc = e
                print(f"⚠️ Attempt {attempt} WS connect failed for {uuid_}: {e}")
                await asyncio.sleep(RETRY_DELAY)
        print(f"❌ WS connection ultimately failed for {uuid_}: {last_exc}")
        return None

    async def add_user(self, session, email):
        token, uuid_ = await self.login_user(session, email)
        if not token or not uuid_:
            return
        async with self.sem:
            ws = await self.connect_ws(token, uuid_)
        if ws:
            user = {"email": email, "token": token, "uuid": uuid_, "ws": ws}
            self.users.append(user)
            asyncio.create_task(self.listen_ws(user))

    async def listen_ws(self, user):
        ws = user["ws"]
        try:
            async for message in ws:
                self.message_count += 1
                print(f"📨 {user['uuid']} received: {message}")
        except Exception as e:
            print(f"🚫 WS closed for {user['uuid']}: {e}")

    async def send_messages(self):
        async with aiohttp.ClientSession() as session:
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
                async with session.post(
                    f"{BACKEND_URL}/send", json=payload, headers=headers
                ) as resp:
                    if resp.status != 201:
                        print(f"❌ Send failed {u1['uuid']} -> {u2['uuid']} ({resp.status})")

    async def cleanup_users(self):
        async with aiohttp.ClientSession() as session:
            async with session.delete(f"{BACKEND_URL}/users") as resp:
                print(f"🧹 Cleanup status: {resp.status}")

    async def run(self):
        emails = [self.generate_email() for _ in range(NUM_USERS)]
        async with aiohttp.ClientSession() as session:
            # register all users
            await asyncio.gather(*[self.register_user(session, email) for email in emails])
            # login and connect WS with throttling
            await asyncio.gather(*[self.add_user(session, email) for email in emails])

        print(f"✅ Total WS connections: {len(self.users)}")
        await self.send_messages()

if __name__ == "__main__":
    tester = LoadTester()
    try:
        asyncio.run(tester.run())
    except KeyboardInterrupt:
        print("🛑 Interrupted, cleaning up...")
        asyncio.run(tester.cleanup_users())
