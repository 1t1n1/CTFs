#!/usr/bin/env python3

import sys
import subprocess
from pathlib import Path
from itertools import product

def generate_binary_strings(n):
    """Generate all binary strings of length n."""
    for bits in product('01', repeat=n):
        yield ''.join(bits)

def main():
    if len(sys.argv) != 4:
        print("Usage: python3 run_until_correct_stdin.py <program_path> <length> <output_file>")
        print("   - Input is sent via STDIN")
        print("   - Stops when 'Wrong' is NOT in stdout")
        sys.exit(1)
    
    program_path = sys.argv[1]
    try:
        length = int(sys.argv[2])
    except ValueError:
        print("Length must be an integer.")
        sys.exit(1)
    
    output_file = Path(sys.argv[3])
    output_file.write_text('')

    total = 2 ** length
    print(f"Testing all {total:,} binary strings of length {length} via STDIN...")
    print(f"Program: {program_path}")
    print(f"Output file: {output_file}")
    print("Stopping when 'Wrong' is NOT in stdout...\n")

    found_correct = False
    count = 0

    try:
        for binary_str in generate_binary_strings(length):
            binary_str = "ICC{" + binary_str + "}"
            count += 1

            try:
                result = subprocess.run(
                    [program_path],
                    input=binary_str,
                    capture_output=True,
                    text=True,
                    timeout=10
                )
            except subprocess.TimeoutExpired:
                stdout = "[TIMEOUT]"
                stderr = ""
            except Exception as e:
                stdout = f"[EXCEPTION: {e}]"
                stderr = ""
            else:
                stdout = result.stdout.strip()
                stderr = result.stderr.strip()

            output_line = f"{binary_str}: {stdout}"
            if stderr:
                output_line += f" [STDERR: {stderr}]"

            with output_file.open('a', encoding='utf-8') as f:
                f.write(output_line + '\n')

            if "Wrong" not in stdout:
                print(f"\nSUCCESS FOUND at combination #{count:,}")
                print(f"Input (via stdin): {binary_str}")
                print(f"Output: {stdout}")
                if stderr:
                    print(f"STDERR: {stderr}")
                found_correct = True
                break

            if count <= 10 or count % max(1, total // 100) == 0:
                status = "Wrong found" if "Wrong" in stdout else "No 'Wrong'"
                print(f"[{count:,}/{total:,}] {binary_str} → {status}")

    except KeyboardInterrupt:
        print("\nInterrupted by user.")
    except Exception as e:
        print(f"Unexpected error: {e}")

    if found_correct:
        print(f"\nCorrect input found! Results saved to: {output_file}")
    else:
        print(f"\nNo success after {count:,} attempts.")

if __name__ == "__main__":
    main()
