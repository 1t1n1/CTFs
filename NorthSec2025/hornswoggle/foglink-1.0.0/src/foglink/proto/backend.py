from __future__ import annotations

import logging
from dataclasses import dataclass

from foglink.proto.interface import Interface
from foglink.proto.messages import (
    Command,
    Message,
    Password,
    StartUp,
    Terminate,
)

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class Backend(Interface):
    async def receive(self) -> Message:
        msg_type, msg_bytes = await self._receive()
        match msg_type:
            case StartUp.identifier:
                msg = StartUp
            case Password.identifier:
                msg = Password
            case Terminate.identifier:
                msg = Terminate
            case Command.identifier:
                msg = Command
            case _:
                raise TypeError(f"Received unexpected message type: {chr(msg_type[0])}")
        logger.info(f"[RECV] {msg.decode(msg_bytes)}")
        return msg.decode(msg_bytes)

    async def send(self, msg: Message):
        logger.info(f"[SEND] {msg}")
        await self._send(msg.encode())
