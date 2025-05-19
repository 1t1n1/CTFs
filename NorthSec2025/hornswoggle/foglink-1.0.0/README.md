
# Installation
```
python3.12 -m venv env
. env/bin/activate
pip install .
```

# Usage
Here is a sample client and server with a test command:

## Server
```python
import asyncio

from foglink.proto.messages import Command

from foglink.server import ServerApplication, User, create_handler


class TestApplication(ServerApplication):
    async def handle_commands(self, cmd: Command):
        if reply := await super().handle_commands(cmd):
            return reply
        match cmd:
            case Command(name="ECHO"):
                print(f"ECHO command received with {cmd.parameters}")
                return cmd.parameters


async def main(host, port):
    users = {
        "admin": User(name="admin", password="admin", is_admin=True),
        "user": User(name="user"),
    }
    handler = create_handler(TestApplication, users)
    async with await asyncio.start_server(handler, host, port) as server:
        await server.serve_forever()


if __name__ == "__main__":
    asyncio.run(main("127.0.0.1", 8888))
```

## Client
```python
import asyncio
from foglink.client import ClientApplication, create_client


class TestClient(ClientApplication):
    async def echo(self, parameters) -> str:
        return await self.send_command(name="ECHO", **parameters)


async def main(host, port, username, password=""):
    client = create_client(TestClient)
    async with client(host, port) as session:
        await session.connect(username, password)
        reply = await session.echo({"A": 1})
    print("Server returned:", reply)


if __name__ == "__main__":
    asyncio.run(main("127.0.0.1", 8888, "user"))
```

## Command-Line Interface (CLI)
The CLI can be used to interact with the server
```
foglink-cli --host <host> --port <port>
> PING
```