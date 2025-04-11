import asyncio
import aiohttp

BACKEND_URL = "http://192.168.1.64:8080"
NUM_USERS = 4000  # Adjust based on your needs

async def create_user(session, user_id):
    email = f"loadtest-{user_id}@example.com"
    password = "testpassword123"

    async with session.post(
        f"{BACKEND_URL}/register",
        json={"email": email, "password": password}
    ) as response:
        if response.status == 201:
            print(f"Created user {user_id}")
            return {"email": email, "password": password}
        return None

async def main():
    async with aiohttp.ClientSession() as session:
        tasks = [create_user(session, i) for i in range(NUM_USERS)]
        users = await asyncio.gather(*tasks)
        
        # Save users to file for cleanup
        with open("test_users.txt", "w") as f:
            for user in users:
                if user:
                    f.write(f"{user['email']}\n")

if __name__ == "__main__":
    asyncio.run(main())