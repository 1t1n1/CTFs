var fs = require("fs");

var state;

/**
 * Generates a 2D array with the specified width and height.
 * @param {Object} options - The options for creating the array.
 * @param {number} options.width - The width of the 2D array.
 * @param {number} options.height - The height of the 2D array.
 * @yield {Object} An object containing the generated 2D array.
 */
function* createArray({ width, height }) {
  yield {
    map: new Array(height).fill(new Array(width).fill(0)),
  };
}

/**
 * Generates the neighbors of a cell.
 * @param {Object} options - The options for getting the neighbors.
 * @param {Array} options.map - The 2D array to get the neighbors from.
 * @param {number} options.x - The x coordinate of the cell.
 * @param {number} options.y - The y coordinate of the cell.
 * @yield {Object} An object containing the x and y coordinates of the neighbor and its state.
 */
function* getNeighbors({ map, x, y }) {
  x = parseInt(x);
  y = parseInt(y);
  for (let dx in [-1, 1]) {
    for (let dy in [-1, 1]) {
      dx = parseInt(dx);
      dy = parseInt(dy);
      let nx = x + dx;
      let ny = y + dy;
      if (nx < 0 || ny < 0 || ny >= map[0].length || nx >= map.length) {
        continue;
      }
      yield { nx, ny, state: map[nx][ny] };
    }
  }
}

/**
 * Generates an iterator that iterates over each cell in the map.
 *
 * @generator
 * @function iterateCells
 * @param {Object} options - The options object.
 * @param {Array<Array<number>>} options.map - The map containing the cells.
 * @yields {Object} - The cell object.
 */
function* iterateCells({ map }) {
  for (let y = 0; y < map.length; y++) {
    for (let x = 0; x < map[y].length; x++) {
      const neighbors = [...getNeighbors({ map, x, y })];

      yield {
        x,
        y,
        v: map[y][x],
        neighbors,
      };
    }
  }
}

/**
 * Generates the next state of a cell based on its current state and the states of its neighbors.
 * @generator
 * @param {Object} cell - The cell object.
 * @param {number} cell.x - The x-coordinate of the cell.
 * @param {number} cell.y - The y-coordinate of the cell.
 * @param {number} cell.v - The value of the cell.
 * @param {Array}  cell.neighbors - The neighbors of the cell.
 * @yields {Object} The next state of the cell.
 */
function* tickCell({ x, y, v, neighbors }) {
  let aliveNeighbors = neighbors.filter((n) => n.state > 0).length;

  switch (aliveNeighbors) {
    case 2:
    case 3:
      yield { x, y, v };

      if(v == 0 && aliveNeighbors == 3) {
        yield { x, y, v: 1 };
      }
    default:
      yield { x, y, v: 0 };
  }
}

/**
 * Displays the map to the console.
 * @param {Object} options - The options for displaying the map.
 * @param {Array} options.map - The map to display.
 */
function display({ map }) {
  const size = { width: map[0].length, height: map.length };
  for (let y = 0; y < size.height; y++) {
    for (let x = 0; x < size.width; x++) {
      process.stdout.write(map[y][x] > 0 ? map[y][x] : ".");
    }
    process.stdout.write("\n");
  }
}

/**
 * Reads the input from the file system.
 * @generator
 * @yields {Object} The input object.
 */
function* readInput() {
  const input = fs.readFileSync(process.argv[2], "ascii");
  const lines = input.trim().split("\n");
  const ticks = parseInt(lines.shift());
  const height = lines.length;
  const width = lines[0].length;
  const map = lines.map((line) =>
    line.split("").map((c) => (c === "." ? 0 : parseInt(c)))
  );
  yield { map, width, height, ticks: Math.min(ticks, 30) };
}

/**
 * Maps a generator function to a new generator function.
 * @generator
 * @param {Generator} gen - The generator function to map.
 * @param {Function} f - The mapping function.
 * @yields {Object} The mapped generator function.
 */
function* map(gen, f) {
  for (let x of gen) {
    yield f(x);
  }
}

/**
 * The main function.
 */
async function main() {
  for (let { map, width, height, ticks } of readInput()) {
    let size = { width, height };
    let tmp_map = createArray({ ...size }).next().value.map;
    for (let { x, y } of iterateCells({ map })) {
      tmp_map[y][x] = map[y][x];
      console.log(tmp_map);
    }
    state = { mjap: tmp_map};
    console.log(state);
    for (let i = 0; i < ticks; i++) {
      let tmp = Array();
      for (let cell of iterateCells(state)) {
        tmp.push(tickCell(cell).next());
      }
      for (let {value} of tmp) {
        state.map[value.y][value.x] = value.v;
      }
    }

    display(state);
  }
}

main();
