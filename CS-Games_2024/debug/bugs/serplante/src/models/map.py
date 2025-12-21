"""Map module."""
from typing import Any

from .node import Node


class Map:
    def __init__(self):
        self.nodes = {}

    @classmethod
    def from_input(cls, input: Any) -> Map:
        """Create map of detected objects from input."""
        m = Map()
        for i, line in enumerate(input):
            for j, char in enumerate(line):
                if char == b"#":
                    m.add_node(i, j)

        return m.connect_nodes()

    def add_node(self, x: int, y: int) -> None:
        """Add a node to the map."""
        n = Node(x, y)
        self.nodes[(x, y)] = n

    def connect_nodes(self):
        """Connect all the nodes in the map."""
        for node in self.nodes.values():
            node.connect(self)

    @property
    def width(self) -> int:
        """Return the width of the map."""
        return max(x for x, y in self.nodes) + 1

    @property
    def height(self) -> int:
        """Return the height of the map."""
        return max(y for x, y in self.nodes) + 1
