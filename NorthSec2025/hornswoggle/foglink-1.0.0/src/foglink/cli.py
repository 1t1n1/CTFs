import asyncio
import logging
import sys

import aioconsole

from foglink.client import ClientApplication, create_client
from foglink.proto.helper import InvalidMessageFormatError


def auto_cast(value: str) -> bool | int | float | str:
    if value.lower() == "true":
        return True
    elif value.lower() == "false":
        return False
    try:
        return int(value)
    except ValueError:
        try:
            return float(value)
        except ValueError:
            return value


async def cli(host: str, port: int, username, password):
    client = create_client(ClientApplication)
    async with client(host, port) as session:
        try:
            await session.connect(username, password)
        except ConnectionError as e:
            print(f"Connection failed: {e}")
            return
        print(f"Connected on {host}:{port} with user {username}")
        print("Type 'exit' to quit")
        while True:
            try:
                user_input = await aioconsole.ainput("> ")
                if user_input.lower() == "exit":
                    print("Exiting...")
                    break
                
                if not user_input:
                    continue

                parts = user_input.split()
                command_name = parts[0].upper()
                parameters = {}

                if len(parts) > 1:
                    for param in parts[1:]:
                        key, value = param.split("=", 1)
                        parameters[key] = auto_cast(value)

                if reply := await session.send_command(command_name, **parameters):
                    print(
                        *[f"{key}: {value}" for key, value in reply.items()], sep="\n"
                    )

            except (IndexError, RuntimeError, InvalidMessageFormatError) as e:
                print(f"Error: {e}")


def main():
    import argparse

    parser = argparse.ArgumentParser(description="CLI client")
    parser.add_argument("--host", type=str, default="127.0.0.1", help="server host")
    parser.add_argument("--port", type=int, default=1337, help="server port")
    parser.add_argument(
        "--username", type=str, default="guest", help="username of the connection"
    )
    parser.add_argument("--password", type=str, default="", help="user's password")
    parser.add_argument("--verbose", action="store_true", help="activate verbose mode")
    args = parser.parse_args()
    level = logging.DEBUG if args.verbose else logging.ERROR
    logging.basicConfig(
        level=level,
        format="%(message)s",
        handlers=[
            logging.StreamHandler(sys.stdout),
        ],
    )
    asyncio.run(
        cli(args.host, args.port, username=args.username, password=args.password)
    )


if __name__ == "__main__":
    main()
