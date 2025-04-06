import asyncio
import aiohttp

BACKEND_URL = "http://localhost:8080"

async def delete_user(session, email):
    async with session.delete(
        f"{BACKEND_URL}/users/{email}",  # Update to your delete endpoint
        headers={"Authorization": "Bearer ADMIN_TOKEN"}  # Add admin auth
    ) as response:
        if response.status == 200:
            print(f"Deleted {email}")

async def main():
    async with aiohttp.ClientSession() as session:
        with open("test_users.txt") as f:
            emails = [line.strip() for line in f.readlines()]
        
        tasks = [delete_user(session, email) for email in emails]
        await asyncio.gather(*tasks)

if __name__ == "__main__":
    asyncio.run(main())