#!/usr/bin/python3

import platform
import random
import time
from datetime import datetime

SLEEP_DELAY = 10

MIN_NUMBER = 0
MAX_NUMBER = 54321


def main():
    print("LITTLE くコ:彡 GAME")
    print(f"Made with Python {platform.python_version()}")
    random.seed(datetime.now().microsecond)
    game_counter = 1
    while True:
        print(f"*********************************")
        print(f" GAME #{game_counter}")
        print(f"*********************************")

        number = random.randint(MIN_NUMBER, MAX_NUMBER)
        print(number)

        print(f"I've chosen a number. You have 1 attempt to guess it.")

        while True:
            try:
                print(f"Enter your guess: ", end="")
                guess = int(input())
                break
            except ValueError:
                print("Invalid input. Enter an integer!")

        if guess == number:
            print("You guessed right!")
            with open('flag.txt', "r") as f:
                flag = f.read()
                print(flag)
            return
        else:
            print("Your guess was incorrect!")

        print(f"The number was: {number}")
        print("Better luck next time!")
        print(f"Little くコ:彡 will take a {SLEEP_DELAY} seconds nap...")
        time.sleep(SLEEP_DELAY)
        game_counter += 1


if __name__ == "__main__":
    main()
