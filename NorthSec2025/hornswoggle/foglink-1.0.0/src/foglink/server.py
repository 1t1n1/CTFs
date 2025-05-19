from __future__ import annotations

import asyncio
import logging
from dataclasses import dataclass, field
from typing import Callable, Type

from foglink.proto.backend import Backend
from foglink.proto.helper import ValuesDict
from foglink.proto.messages import (
    AuthenticationOk,
    Command,
    CommandComplete,
    CommandReply,
    ErrorResponse,
    ErrorSeverity,
    InvalidMessageFormatError,
    Password,
    PasswordRequest,
    Ready,
    StartUp,
    Terminate,
)

logger = logging.getLogger(__name__)


@dataclass
class User:
    name: str
    password: str | None = None
    is_admin: bool = False


@dataclass
class ServerApplication:
    backend: Backend
    users: list[User] = field(default_factory=lambda: [User(name="guest")])
    current_user: User | None = field(default=None, init=False)

    async def handle_commands(self, cmd: Command) -> ValuesDict:
        """
        This method can be extended to handle additional commands.
        """
        match cmd:
            case Command(name="PING"):
                return {"success": True}
            case Command(name="GET_CURRENT_USER"):
                return {
                    "success": True,
                    "username": self.current_user.name,
                    "is_admin": self.current_user.is_admin,
                }

    async def run(self):
        try:
            if not await self.handle_startup():
                logger.error("Connection failed")
                return
        except (
            InvalidMessageFormatError,
            TypeError,
            ValueError,
        ) as e:
            logger.error(f"Error processing StartUp message: {e}")
            await self.backend.send(ErrorResponse(message=str(e)))
            return

        await self.backend.send(Ready())

        while True:
            try:
                msg = await self.backend.receive()
                match msg:
                    case Command(name=name):
                        if reply := await self.handle_commands(msg):
                            await self.backend.send(CommandReply(reply))
                        await self.backend.send(CommandComplete(name))
                    case Terminate():
                        return
                    case _:
                        await self.backend.send(
                            ErrorResponse(
                                message=f"Received unsupported message type: {type(msg)}"
                            )
                        )
            except Exception as e:
                logger.exception(e)
                await self.backend.send(ErrorResponse(message=str(e)))

    async def handle_startup(self) -> bool:
        msg = await self.backend.receive()
        if type(msg) is not StartUp:
            await self.backend.send(
                ErrorResponse(
                    severity=ErrorSeverity.CRITICAL,
                    message=f"Excepted StartUpMessage, received {type(msg)}",
                )
            )
            return False
        username = msg.parameters["username"]
        if username not in self.users:
            await self.backend.send(
                ErrorResponse(message=f"Username {username} not found")
            )
            return False
        self.current_user = self.users[username]
        if self.current_user.password is None:
            await self.backend.send(AuthenticationOk())
            return True
        await self.backend.send(PasswordRequest())
        msg = await self.backend.receive()
        if type(msg) is not Password:
            await self.backend.send(
                ErrorResponse(
                    severity=ErrorSeverity.CRITICAL,
                    message=f"Excepted PasswordMessage, received {type(msg)}",
                )
            )
            raise RuntimeError()
        if msg.password != self.current_user.password:
            await self.backend.send(ErrorResponse(message="Invalid password"))
            return False
        await self.backend.send(AuthenticationOk())
        return True


def create_handler[T: ServerApplication](
    app: Type[T], users=None
) -> Callable[[asyncio.StreamReader, asyncio.StreamWriter], None]:
    users = users or {}

    async def handler(reader: asyncio.StreamReader, writer: asyncio.StreamWriter):
        addr = writer.get_extra_info("peername")
        logger.info(f"Connection by {addr[0]}:{addr[1]}")
        backend = Backend(reader, writer)
        try:
            await app(backend, users).run()
        except asyncio.IncompleteReadError:
            pass
        finally:
            logger.info(f"Client {addr[0]}:{addr[1]} disconnected.")
            await backend.close()

    return handler
