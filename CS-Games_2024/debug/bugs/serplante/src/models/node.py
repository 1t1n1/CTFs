"""
Node in a 2d grid.
"""

from .map import Map


class Node:
    def __init__(self, x: int, y: int):
        self.x = x
        self.y = y
        self.connected_nodes = []

    def __repr__(self) -> str:
        return f"N({self.x}, {self.y})"

    def tuple(self) -> (int, int):
        return (self.x, self.y)

    def connect(self, map: Map):
        """
        Connect to other nodes in the map.
        """
        for i in range(
            -1,
            1,
        ):
            for j in range(
                -1,
                1,
            ):
                if i == 0 and j == 0:
                    continue

                if (
                    self.x + i >= 0
                    and self.x + i < map.width
                    and self.y + j >= 0
                    and self.y + j < map.height
                ):
                    p = (
                        self.x + i,
                        self.y + j,
                    )

                    if p in map.nodes:
                        self.connected_nodes.append(
                            map.nodes[
                                self.x + i,
                                self.y + j,
                            ]
                        )
