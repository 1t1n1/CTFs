from __future__ import annotations

import logging
from dataclasses import dataclass

from foglink.proto.interface import Interface
from foglink.proto.messages import (
    AuthenticationOk,
    CommandComplete,
    CommandReply,
    ErrorResponse,
    Message,
    PasswordRequest,
    Ready,
)

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class Frontend(Interface):
    async def receive(self) -> Message:
        msg_type, msg_bytes = await self._receive()
        match msg_type:
            case AuthenticationOk.identifier:
                msg = AuthenticationOk
            case PasswordRequest.identifier:
                msg = PasswordRequest
            case Ready.identifier:
                msg = Ready
            case ErrorResponse.identifier:
                msg = ErrorResponse
            case CommandReply.identifier:
                msg = CommandReply
            case CommandComplete.identifier:
                msg = CommandComplete
            case _:
                raise TypeError(f"Received unexpected message type: {chr(msg_type[0])}")
        logger.info(f"[RECV] {msg.decode(msg_bytes)}")
        return msg.decode(msg_bytes)

    async def send(self, msg: Message):
        logger.info(f"[SEND] {msg}")
        await self._send(msg.encode())
