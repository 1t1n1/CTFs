import requests
import json

DEBUG = True

URL = "http://challenges.ulctf.ca:30202/maze"

ROAD = 0
WALL = 1

class MazeSolver:
    def __init__(self):
        # Note: The 2D array that I receive is build so that the 2nd dimension arrays 
        # are the rows
        try:
            response = requests.get(URL, timeout=10)
            if response.status_code == 200:
                if DEBUG:
                    print("hello from the server \U0001F600 \n")
                maze = response.text
                mazeobj = json.loads(maze)
                mazeuni = mazeobj["maze"]
                if DEBUG:
                    print(mazeuni)
                    print(f"Pour avoir la liste rendez-vous ici -->{URL}")
                mazeobj['maze_array']
                self.graphical_maze = mazeobj['maze']
                self.maze = mazeobj['maze_array']
                self.stack = []
                self.visited_tiles = []
            else:
                print("Failed to get response from server.")
        except requests.exceptions.RequestException as e:
            print("Exception occurred:", e)
            print("Please check your network/firewall settings.")

    def getEntry(self):
        for i in range(len(self.maze)):
            if self.maze[i][0] == ROAD:
                return (0, i)

    def getExit(self):
        for i in range(len(self.maze)):
            if self.maze[i][len(self.maze[0]) - 1] == ROAD:
                return (len(self.maze) - 1, i)

    def getNeighbors(self, tile):
        neighbors = []
        for (i, j) in [(1, 0), (0, 1), (0, -1), (-1, 0)]:
            neighborX = tile[0] + i
            neighborY = tile[1] + j
            if ((0 <= neighborX and neighborX < len(self.maze[0])) and
                (0 <= neighborY and neighborY < len(self.maze)) and 
                (i != j) and (self.maze[neighborY][neighborX] == ROAD)):
                neighbors.append((neighborX, neighborY))
        return neighbors

    def send_solution(self):
        print("Solution is")
        print(self.stack)
        data = {
            "path": self.stack
        }
        headers = {
            "Content-Type": "application/json"
        }
        response = requests.post(URL, data=json.dumps(data), headers=headers)
        print(response.text)
        exit()

    def dfs(self, current_tile):
        self.visited_tiles.append(current_tile)
        for neighbor in self.getNeighbors(current_tile):
            if (neighbor == self.getExit()):
                self.stack.append(neighbor)
                self.send_solution()
            elif (not neighbor in self.visited_tiles):
                self.stack.append(neighbor)
                self.dfs(neighbor)
                self.stack.pop()

    def solve(self):
        self.stack.append(self.getEntry())
        self.dfs(self.stack[0])

if __name__ == '__main__':
    solver = MazeSolver()
    solver.solve()