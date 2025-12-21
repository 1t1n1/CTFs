from __future__ import annotations

import asyncio
import logging
from contextlib import asynccontextmanager
from dataclasses import dataclass, field
from typing import AsyncContextManager, AsyncGenerator, Callable, Type

from foglink.proto.frontend import Frontend
from foglink.proto.helper import ValuesDict
from foglink.proto.messages import (
    AuthenticationOk,
    Command,
    CommandComplete,
    CommandReply,
    ErrorResponse,
    Password,
    PasswordRequest,
    Ready,
    StartUp,
    Terminate,
)

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class ClientApplication:
    frontend: Frontend
    lock: asyncio.Lock = field(default_factory=asyncio.Lock, init=False)

    async def connect(self, username, password):
        await self.frontend.send(StartUp(parameters={"username": username}))
        while True:
            msg = await self.frontend.receive()
            match msg:
                case AuthenticationOk():
                    pass
                case PasswordRequest():
                    await self.frontend.send(Password(password))
                case Ready():
                    return
                case ErrorResponse():
                    raise ConnectionError(msg.message)
                case _:
                    raise ConnectionError(f"Unexpected message received: {msg}")

    async def close(self):
        try:
            await self.frontend.send(Terminate())
        finally:
            await self.frontend.close()

    async def send_command(
        self, name: str, **parameters: ValuesDict
    ) -> ValuesDict | None:
        async with self.lock:
            await self.frontend.send(Command(name=name, parameters=parameters))
            reply = None
            while True:
                msg = await self.frontend.receive()
                match msg:
                    case CommandReply():
                        reply = msg
                    case CommandComplete(name=name):
                        return reply.values if reply else None
                    case ErrorResponse():
                        raise RuntimeError(msg.message)
                    case _:
                        raise RuntimeError(f"Unexpected message received: {msg}")

    async def ping(self) -> None:
        return await self.send_command(name="PING")
    
    async def get_current_user(self) -> None:
        return await self.send_command(name="GET_CURRENT_USER")


def create_client[T: ClientApplication](
    app: Type[T],
) -> Callable[[str, int], AsyncContextManager[T]]:
    @asynccontextmanager
    async def client(host: str, port: int) -> AsyncGenerator[T, None]:
        reader, writer = await asyncio.open_connection(host, port)
        frontend = Frontend(reader, writer)
        client = app(frontend)
        try:
            yield client
        finally:
            await client.close()

    return client
