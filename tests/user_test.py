#!/usr/bin/env python3
import base64
import time
import requests
import uuid
import random
import string
import os
import jwt   # pip install PyJWT
from dotenv import load_dotenv # pip install python-dotenv
HOST = "127.0.0.1"
PORT = 8080
BASE_URL = f"http://{HOST}:{PORT}"
# Load .env into os.environ
load_dotenv()  
# JWT_SECRET must match the base64-encoded JWT_SECRET your Go service uses
JWT_SECRET = os.environ.get("JWT_SECRET", "")

def random_email():
    return f"{uuid.uuid4().hex[:8]}@example.com"

def random_password(length=12):
    chars = string.ascii_letters + string.digits
    return ''.join(random.choice(chars) for _ in range(length))

def register_user(email, password):
    return requests.post(f"{BASE_URL}/register", json={"email": email, "password": password})

def login_user(email, password):
    return requests.post(f"{BASE_URL}/login", json={"email": email, "password": password})

def send_message(token, receiver, content):
    headers = {"Authorization": f"Bearer {token}"}
    return requests.post(f"{BASE_URL}/send", json={"receiver": receiver, "content": content}, headers=headers)

def forgot_password(email):
    return requests.post(f"{BASE_URL}/reset-password", json={"email": email})

# Hypothetical “confirm reset” endpoint—adjust path when you implement it
def reset_password(token, new_password):
    return requests.post(f"{BASE_URL}/reset-password/confirm", json={"token": token, "password": new_password})

def generate_reset_token(user_id, expires_in=15*60):
    """
    Generate a JWT reset token with typ="pwd_reset" and the same secret your server uses.
    """
    secret = base64.b64decode(JWT_SECRET)
    payload = {
        "sub": user_id,
        "exp": int(time.time()) + expires_in,
        "typ": "pwd_reset"
    }
    return jwt.encode(payload, secret, algorithm="HS256")

def main():
    print("=== 1. Register with invalid data ===")
    r = register_user("not-an-email", "short")
    print(r.status_code, r.json(), "\n")

    print("=== 2. Login with incorrect credentials ===")
    r = login_user("noone@example.com", "wrongpass")
    print(r.status_code, r.text, "\n")

    # === Set up two valid users ===
    sender_email = random_email()
    sender_pass  = random_password()
    recv_email   = random_email()
    recv_pass    = random_password()

    # Register sender
    r = register_user(sender_email, sender_pass)
    print("Register sender:", r.status_code, r.json())
    # Register receiver
    r = register_user(recv_email, recv_pass)
    print("Register receiver:", r.status_code, r.json(), "\n")

    print("=== 3. Login with correct credentials ===")
    r = login_user(sender_email, sender_pass)
    print(r.status_code, r.json(), "\n")
    token       = r.json().get("token")
    user_id     = r.json().get("uuid")
    recv_id     = login_user(recv_email, recv_pass).json().get("uuid")

    print("=== 4. Send message with incorrect token ===")
    r = send_message("INVALID_TOKEN", recv_id, "Hello")
    print(r.status_code, r.text, "\n")

    print("=== 5. Send message with correct token ===")
    r = send_message(token, recv_id, "Hi there!")
    print(r.status_code, r.json(), "\n")

    print("=== 6. Forgot password with incorrect email ===")
    r = forgot_password("doesnotexist@example.com")
    print(r.status_code, r.json(), "\n")

    print("=== 7. Forgot password with correct email ===")
    r = forgot_password(sender_email)
    print(r.status_code, r.json(), "\n")

    print("=== 8. Reset password with incorrect token ===")
    r8 = reset_password("badtoken", "NewPass123!")
    print(r8.status_code, r8.text, "\n")

    print("=== 9. Reset password with correct token ===")
    valid_token = generate_reset_token(user_id)
    r9 = reset_password(valid_token, "NewPass123!")
    print(r9.status_code, r9.json(), "\n")

if __name__ == "__main__":
    print(f"=== JWT_SECRET === ${JWT_SECRET}")
    main()