from __future__ import annotations

import asyncio
import logging
from dataclasses import dataclass
from typing import Tuple

from foglink.proto.messages import MessageSize

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class Interface:
    reader: asyncio.StreamReader
    writer: asyncio.StreamWriter

    async def _receive(self) -> Tuple[bytes, bytes]:
        header = await self.reader.readexactly(1 + MessageSize)
        if len(header) != 1 + MessageSize:
            raise ValueError(
                f"Invalid header length. Received: {len(header)} bytes, expected {MessageSize + 1}"
            )
        msg_length = int.from_bytes(header[1:], byteorder="big")
        msg_body = await self.reader.readexactly(msg_length)
        logger.debug(f"Byte received: {header + msg_body}")
        return header[0:1], msg_body

    async def _send(self, msg: bytes) -> None:
        logger.debug(f"Byte sent: {msg}")
        self.writer.write(msg)
        await self.writer.drain()

    def is_closed(self):
        return self.writer.is_closing()

    async def close(self):
        if self.is_closed():
            return
        self.writer.close()
        await self.writer.wait_closed()
