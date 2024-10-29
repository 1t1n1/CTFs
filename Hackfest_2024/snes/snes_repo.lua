-- base64 decode
local b64chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/'
local b64map = {}
for i = 1, #b64chars do
    b64map[b64chars:sub(i, i)] = i - 1
end

local function base64_decode(data)
    data = data:gsub('[^' .. b64chars .. '=]', '')
    return (data:gsub('.', function(x)
        if x == '=' then return '' end
        local r, f = '', (b64map[x] or 0)
        for i = 6, 1, -1 do
            r = r .. (f % 2 ^ i - f % 2 ^ (i - 1) > 0 and '1' or '0')
        end
        return r
    end):gsub('%d%d%d?%d?%d?%d?%d?%d?', function(x)
        if (#x ~= 8) then return '' end
        local c = 0
        for i = 1, 8 do c = c + (x:sub(i, i) == '1' and 2 ^ (8 - i) or 0) end
        return string.char(c)
    end))
end

-- Define the Game class and items
Game = {}
Game.__index = Game

-- Constructor for the Game class
function Game:new(title, serial, ROMsize)
    local instance = setmetatable({}, Game)
    instance.title = title or "Unknown"
    instance.serial = serial or "Unknown"
    instance.ROMsize = ROMsize or "Unknown"
    
    return instance
end

-- Verify the repository file exists
filepath = "repository.txt"
file = io.open (filepath, "r+")
if not file then
	file = io.open (filepath, "w")
	io.close(file)
	file = io.open (filepath, "r+")
end

io.output(file)

function Game:printDetails()
    io.write("\n-- Title: " .. self.title .. "\n-- Serial: " .. self.serial .. "\n-- ROMsize: " .. self.ROMsize .. "\n------------------------------------------------")
end

local Game1 = Game:new("Uncharted Waters - New Horizons", "SNS-QL-USA", "2 MB")
local Game2 = Game:new("Donkey Kong Country 2 - Diddy's Kong Quest", "SNS-ADNE-USA", "4 MB")
local Game3 = Game:new("Power Rangers Zeo - Battle Racers", "SNSP-A4RP-EUR", "1 MB")
local Game4 = Game:new("Final Fanstasy II", "SNS-F4-USA", "1 MB")
local Game5 = Game:new("DinoCity", "SNS-DW-USA", "1 MB")

-- Print repository
io.write("----------------SNES repository-----------------")
Game1:printDetails()
Game2:printDetails()
Game3:printDetails()
Game4:printDetails()
Game5:printDetails()
io.close(file)

file = io.open (filepath, "r")
print(file:read("*a"))
io.close(file)

-- DEBUG statement, don't forget to remove!!1 ^w^
local sandbox_env = {
        dofile = dofile
}

-- Function to execute decoded bytecode
local function execute_bytecode(encoded_data)
    -- Decode Base64 data
    local bytecode = base64_decode(encoded_data)

    local func, err = load(bytecode, nil, 'b', sandbox_env)
    if not func then
        print("Failed to load bytecode:", err)
        return false
    else
        -- Load the new entry
        local success, title, serial, romSize = pcall(func)
        
        if not success then
            print("Error during execution:", title)
            return false
        else
            return true, title, serial, romSize
        end
    end
end

-- Main function to handle user input and run bytecode
local function main()
	while true do
	    print("\n---Enter 'exit' anytime to quit and reset the repository.---!")
	    print("\nEnter your base64 encoded object to add to the SNES repository:")
	    local input = io.read("*line")
	    
	    if input == "exit" then
            	print("Exiting...")
            	os.remove(filepath)
            	break  -- Exit the loop
            end

	    -- Execute the bytecode and capture results
	    local success, title, serial, romSize = execute_bytecode(input)

	    -- Check if bytecode executed successfully and validate output
	    if success then
		if title then
		    -- Create a new instance of Game with the results
		    local game = Game:new(title, serial, romSize)
		    file = io.open (filepath, "a")
		    io.output(file)
		    game:printDetails()
		    io.close(file)
		    file = io.open (filepath, "r")
		    print(file:read("*a"))
		    io.close(file)
		else
		    print("There was an error adding you entry...")
		end
	    else
		print("Failed to add entry to repository!")
	    end
	end
end
main()
