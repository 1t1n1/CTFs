"""
Represent a potential snake.
"""

from .node import Node
from typing import Sequence


class Snake:
    def __init__(self, nodes: Sequence[Node]):
        self.nodes = nodes
        self.head = None
        self.tail = None

    def is_valid(self):
        """Determines if the snake is valid."""
        if len(self.nodes) < 5:
            return False

        if self.ends():
            return True

    def ends(self):
        """Return the head and tail of the snake."""
        ends = []

        for node in self.nodes:
            if len(node.connected_nodes) == 1:
                ends.append(node)

        self.head, self.tail = sorted(ends)
        return (self.head, self.tail)
