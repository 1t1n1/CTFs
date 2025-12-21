"""Main."""
import sys
from models import Map
from pathlib import Path
from detector import Detector

if len(sys.argv) < 2:
    print("Usage: python3 serplante.py <input_file>")
    sys.exit(1)

p = str(Path(sys.argv[1]).as_posix())
with open(p) as fd:
    m = Map.from_input(fd)

d = Detector(m)
heads: list["..."] = []
groups = 0
for snake in d.detect():
    groups += 1
    if not snake.is_valid:
        continue
    heads.append(snake.head)

count = len(heads)

print("groups: {groups}")
print("snakes: {count}")
