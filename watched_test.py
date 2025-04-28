import asyncio
import ssl

from aioquic.asyncio import connect
from aioquic.h3.connection import H3_ALPN, H3Connection
from aioquic.h3.events import DataReceived, HeadersReceived
from aioquic.quic.configuration import QuicConfiguration

class HttpClient:
    def __init__(self):
        self._http = None

    async def connect_and_listen(self, host: str, port: int):
        print("[1] Setting up QUIC configuration...")
        configuration = QuicConfiguration(
            alpn_protocols=H3_ALPN,
            is_client=True,
        )
        configuration.verify_mode = ssl.CERT_NONE  # accept self-signed certs

        print(f"[2] Connecting to https://{host}:{port} ...")
        async with connect(
            host=host,
            port=port,
            configuration=configuration,
        ) as client:
            print("[3] Connected! Sending subscription request...")

            self._http = H3Connection(client._quic)

            stream_id = client._quic.get_next_available_stream_id()
            headers = [
                (b":method", b"GET"),
                (b":scheme", b"https"),
                (b":authority", host.encode()),
                (b":path", b"/subscribe"),
            ]
            self._http.send_headers(stream_id, headers)
            self._http.send_data(stream_id, b"", end_stream=True)

            await self.listen_for_updates(client)

    async def listen_for_updates(self, client):
        print("[4] Waiting for updates...")

        try:
            while True:
                event = await client.wait_event()
                if event is None:
                    print("[5] Server closed the connection.")
                    break

                if isinstance(event, HeadersReceived):
                    headers = dict(event.headers)
                    status = headers.get(b":status")
                    if status:
                        status = int(status.decode())
                        if status == 200:
                            print("[✔] Subscription accepted.")
                        else:
                            print(f"[✘] Unexpected HTTP status: {status}")
                            break

                if isinstance(event, DataReceived):
                    data = event.data.decode().strip()
                    if data.startswith("data:"):
                        content = data[5:].strip()
                        print(f"[📂] Update:\n{content}\n")

        except Exception as e:
            print(f"[!] Connection error: {e}")

async def main():
    client = HttpClient()
    await client.connect_and_listen("localhost", 8443)  # <<< corrected port

if __name__ == "__main__":
    print("=== HTTP/3 Watch Test Client ===")
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n[!] Client terminated by user.")