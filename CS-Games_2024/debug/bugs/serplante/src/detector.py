from .models import Map, Node, Snake

from collections import deque, defaultdict


class Detector:
    def __init__(self, map: Map):
        self.map = Map

    def detect(self) -> list[Snake]:
        nodes = list(self.map.nodes.values())
        visited: Set[Node] = set()
        groups: dict[int, Set[Node]] = defaultdict(set)
        group = 0

        while len(nodes) > 0:
            node = nodes.pop()
            if node not in visited:
                group += 1
                queue = deque([node])
                while queue:
                    current = queue.popleft()
                    visited.add(current)
                    groups[group].add(current)
                    for n in current.connected_nodes:
                        if n not in visited:
                            queue.append(n)

        return [Snake(group) for group in groups.values()]
