from __future__ import annotations

from dataclasses import dataclass, field
from enum import IntEnum, auto
from typing import Protocol, Type

from foglink.proto.helper import (
    InvalidMessageFormatError,
    MessageSize,
    ValuesDict,
    decode_values,
    encode_values,
    finalize_message,
    initialize_message,
)

ProtocolVersionNumber = 1


class MessageID:
    StartUp = b"S"
    AuthenticationOk = b"O"
    PasswordRequest = b"P"
    Ready = b"Z"
    Password = b"p"
    ErrorResponse = b"E"
    Terminate = b"X"
    Command = b"C"
    CommandComplete = b"c"
    CommandReply = b"R"


class Message[T](Protocol):
    identifier: bytes

    @classmethod
    def decode(cls: Type[T], data: bytes) -> T: ...

    def encode(self) -> bytes: ...


@dataclass
class StartUp:
    identifier: bytes = field(default=MessageID.StartUp, init=False)
    protocol_version: int = ProtocolVersionNumber
    parameters: ValuesDict = field(default_factory=dict)

    @classmethod
    def decode(cls: Type[StartUp], data: bytes) -> StartUp:
        if len(data) < MessageSize:
            raise ValueError("StartUp message too short")
        protocol_version = int.from_bytes(data[:MessageSize], byteorder="big")
        if protocol_version != ProtocolVersionNumber:
            raise ValueError(f"Invalid protocol version: {protocol_version}")
        parameters = decode_values(data[MessageSize:])
        if "username" not in parameters:
            raise ValueError("Username parameter is required")
        return cls(protocol_version=protocol_version, parameters=parameters)

    def encode(self) -> bytes:
        data = initialize_message(self.identifier)
        data.extend(ProtocolVersionNumber.to_bytes(MessageSize, byteorder="big"))
        data.extend(encode_values(self.parameters))
        return finalize_message(data)


@dataclass
class AuthenticationOk:
    identifier: bytes = field(default=MessageID.AuthenticationOk, init=False)

    @classmethod
    def decode(cls: Type[AuthenticationOk], data: bytes) -> AuthenticationOk:
        return cls()

    def encode(self) -> bytes:
        data = initialize_message(self.identifier)
        return finalize_message(data)


@dataclass
class PasswordRequest:
    identifier: bytes = field(default=MessageID.PasswordRequest, init=False)

    @classmethod
    def decode(cls: Type[PasswordRequest], data: bytes) -> PasswordRequest:
        return cls()

    def encode(self) -> bytes:
        data = initialize_message(self.identifier)
        return finalize_message(data)


@dataclass
class Ready:
    identifier: bytes = field(default=MessageID.Ready, init=False)

    @classmethod
    def decode(cls: Type[Ready], data: bytes) -> Ready:
        return cls()

    def encode(self) -> bytes:
        data = initialize_message(self.identifier)
        return finalize_message(data)


@dataclass
class Password:
    identifier: bytes = field(default=MessageID.Password, init=False)
    password: str = ""

    @classmethod
    def decode(cls: Type[Password], data: bytes) -> Password:
        password_end = data.find(0)
        if password_end < 0:
            raise InvalidMessageFormatError(cls.__name__)
        password = data[:password_end].decode()
        return cls(password=password)

    def encode(self) -> bytes:
        data = initialize_message(self.identifier)
        data.extend(self.password.encode())
        data.extend(b"\x00")
        return finalize_message(data)


class ErrorSeverity(IntEnum):
    INFO = auto()
    ERROR = auto()
    CRITICAL = auto()

    def as_bytes(self) -> bytes:
        return self.value.to_bytes(1, byteorder="big")


@dataclass
class ErrorResponse:
    identifier: bytes = field(default=MessageID.ErrorResponse, init=False)
    message: str
    severity: ErrorSeverity = ErrorSeverity.ERROR

    @classmethod
    def decode(cls: Type[ErrorResponse], data: bytes) -> ErrorResponse:
        try:
            severity = ErrorSeverity(data[0])
        except ValueError:
            raise ValueError(f"Invalid error severity: {data[0]}")
        message_end = data.find(0, 1)
        if message_end < 0:
            raise InvalidMessageFormatError(cls.__name__)
        message = data[1:message_end].decode()
        return cls(severity=severity, message=message)

    def encode(self) -> bytes:
        data = initialize_message(self.identifier)
        data.extend(self.severity.as_bytes())
        data.extend(self.message.encode())
        data.extend(b"\x00")
        return finalize_message(data)


@dataclass
class Terminate:
    identifier: bytes = field(default=MessageID.Terminate, init=False)

    @classmethod
    def decode(cls: Type[Terminate], data: bytes) -> Terminate:
        return cls()

    def encode(self) -> bytes:
        data = initialize_message(self.identifier)
        return finalize_message(data)


@dataclass
class Command:
    identifier: bytes = field(default=MessageID.Command, init=False)
    name: str
    parameters: ValuesDict = field(default_factory=dict)

    @classmethod
    def decode(cls: Type[Command], data: bytes) -> Command:
        name_end = data.find(0)
        name = data[:name_end].decode()
        if type(name) is not str:
            raise ValueError(f"Invalid command name: {name}")
        parameters = decode_values(data[name_end + 1 :])
        return cls(name=name, parameters=parameters)

    def encode(self) -> bytes:
        data = initialize_message(self.identifier)
        data.extend(self.name.encode())
        data.extend(b"\x00")
        data.extend(encode_values(self.parameters))
        return finalize_message(data)


@dataclass
class CommandComplete:
    identifier: bytes = field(default=MessageID.CommandComplete, init=False)
    name: str

    @classmethod
    def decode(cls: Type[CommandComplete], data: bytes) -> CommandComplete:
        name_end = data.find(0)
        name = data[:name_end].decode()
        if type(name) is not str:
            raise ValueError(f"Invalid command name: {name}")
        return cls(name=name)

    def encode(self) -> bytes:
        data = initialize_message(self.identifier)
        data.extend(self.name.encode())
        data.extend(b"\x00")
        return finalize_message(data)


@dataclass
class CommandReply:
    identifier: bytes = field(default=MessageID.CommandReply, init=False)
    values: ValuesDict

    @classmethod
    def decode(cls: Type[CommandReply], data: bytes) -> CommandReply:
        values = decode_values(data)
        return cls(values=values)

    def encode(self) -> bytes:
        data = initialize_message(self.identifier)
        data.extend(encode_values(self.values))
        return finalize_message(data)
