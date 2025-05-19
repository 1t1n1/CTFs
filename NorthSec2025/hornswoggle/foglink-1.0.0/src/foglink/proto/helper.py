import struct
from enum import IntEnum, auto
from typing import Union

type ValueType = Union[int, float, bool, str]
type ValuesDict = dict[str, ValueType]

MessageSize = 1
IntSize = 1

class InvalidMessageFormatError(Exception):
    def __init__(self, message_type: str):
        super().__init__(f"Invalid format for message type: {message_type}")
        self.message_type = message_type


class ValueFlag(IntEnum):
    INT = auto()
    FLOAT = auto()
    BOOL = auto()
    STR = auto()
    DICT = auto()

    def as_bytes(self) -> bytes:
        return self.value.to_bytes(1, byteorder="big")


def decode_values(data: bytes, offset: int = 0) -> ValuesDict:
    values = {}
    length = len(data)
    while offset < length - 1:
        key_end = data.find(0, offset)
        if key_end < 0:
            raise InvalidMessageFormatError("Missing null terminator for key")
        key = data[offset:key_end].decode()
        offset = key_end + 1
        value_type = data[offset]
        offset += 1
        match value_type:
            case ValueFlag.INT:
                # TODO: Support for larger int size
                byte_size = data[offset]
                offset += IntSize
                value = int.from_bytes(
                    data[offset : offset + byte_size], byteorder="big"
                )
                offset += byte_size
            case ValueFlag.FLOAT:
                value = struct.unpack("!f", data[offset : offset + 4])[0]
                offset += 4
            case ValueFlag.BOOL:
                value = bool(data[offset])
                offset += 1
            case ValueFlag.STR:
                str_end = data.find(0, offset)
                if str_end < 0:
                    raise InvalidMessageFormatError(
                        "Missing null terminator for string value"
                    )
                value = data[offset:str_end].decode()
                offset = str_end + 1
            case _:
                raise TypeError(f"Invalid value type received: {value_type}")
        values[key] = value

    if data[-1] != 0:
        raise ValueError("Expected null byte at the end of parameters section")
    return values


def encode_values(values: ValuesDict) -> bytes:
    data = bytearray()
    for key, value in values.items():
        data.extend(key.encode())
        data.extend(b"\x00")
        if isinstance(value, bool):
            value_type = ValueFlag.BOOL
            value_bytes = struct.pack("?", value)
        elif isinstance(value, int):
            data.extend(ValueFlag.INT.to_bytes())
            byte_size = (value.bit_length() + 7) // 8
            # TODO: Support for larger int size
            data.extend((byte_size).to_bytes(IntSize, byteorder="big"))
            data.extend(value.to_bytes(byte_size, byteorder="big"))
            continue
        elif isinstance(value, float):
            value_type = ValueFlag.FLOAT
            value_bytes = struct.pack("!f", value)
        elif isinstance(value, str):
            value_type = ValueFlag.STR
            value_bytes = value.encode()
            value_bytes += b"\x00"
        else:
            raise TypeError(f"Invalid value type received: {type(value)}")
        data.extend(value_type.to_bytes())
        data.extend(value_bytes)

    data.extend(b"\x00")
    return bytes(data)


def initialize_message(identifier: bytes) -> bytearray:
    data = bytearray(MessageSize + 1)
    data[0] = identifier[0]
    return data


def finalize_message(data: bytearray) -> bytes:
    length = len(data) - 1 - MessageSize
    length %= 0xFF**MessageSize
    data[1 : 1 + MessageSize] = length.to_bytes(MessageSize, byteorder="big")
    return bytes(data)
