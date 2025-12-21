-- filename: @vts_obfuscated.lua
-- version: lua53
-- line: [0, 0] id: 0
local r0_0 = require("socket")
local r1_0 = require("ssl")
local r2_0 = require("luasql.sqlite3")
local r3_0 = require("posix.signal")
local r4_0 = require("crc32")
a6 = "\u{1}"
a7 = "\0"
a8 = 100
local r5_0 = {}
local r6_0 = {}
a11 = true

-- Input validation constants
local MAX_STRING_LENGTH = 256
local MAX_VOTE_COUNT = 10000
local MAX_CANDIDATES_PER_BALLOT = 100

a12 = function(r0_1)
  -- line: [29, 31] id: 1
  a11 = false
end
a13 = false
a14 = "              *** Val Verde Central Electoral Commission * Vote Tabulation Service ***              "
a15 = " Press Ctrl+C to exit                                                                               "

-- Input sanitization function
local function sanitize_input(input)
  if type(input) ~= "string" then
    return nil
  end
  if #input > MAX_STRING_LENGTH then
    return nil
  end
  -- Remove any SQL injection attempts and dangerous characters
  local sanitized = string.gsub(input, "['\";%-%-%[%]%(%)%*%+%?%^%$%|%.\\]", "")
  return sanitized
end

-- Prepared statement helper
local function execute_prepared_query(stmt, ...)
  local params = {...}
  for i, param in ipairs(params) do
    if type(param) == "string" then
      params[i] = sanitize_input(param)
      if not params[i] then
        return nil
      end
    end
  end
  return a18:execute(string.format(stmt, table.unpack(params)))
end

a16 = function()
  -- line: [41, 74] id: 2
  if arg[1] == "-s" then
    a13 = true
    print(a14)
    print(a15)
  end
  r3_0.signal(r3_0.SIGINT, a12)
  a17 = r2_0.sqlite3()
  a18 = a17:connect("./vts.db")
  
  if not a18 then
    error("Failed to connect to database")
  end
  
  -- Enable foreign keys and WAL mode for better concurrency
  a18:execute("PRAGMA foreign_keys = ON;")
  a18:execute("PRAGMA journal_mode = WAL;")
  
  local r0_2 = a18:execute("SELECT name FROM voters;")
  if not r0_2 then
    error("Failed to query voters table")
  end
  
  local r1_2 = r0_2:fetch()
  while r1_2 do
    local sanitized_name = sanitize_input(r1_2)
    if sanitized_name then
      r5_0[sanitized_name] = true
    end
    r1_2 = r0_2:fetch()
  end
  r0_2:close()
  
  local r2_2 = nil
  r0_2 = a18:execute("SELECT crc FROM machines;")
  if not r0_2 then
    error("Failed to query machines table")
  end
  
  r2_2 = r0_2:fetch()
  while r2_2 do
    if type(r2_2) == "number" then
      r6_0[r2_2] = true
    end
    local r3_2 = r0_2:fetch()
    r2_2 = r3_2
  end
  r0_2:close()
end

local r7_0 = {
  [0] = 31,
  [1] = 32,
  [2] = 33,
  [3] = 34,
  [4] = 35,
  [5] = 31,
  [6] = 32,
  [7] = 33,
  [8] = 34,
  [9] = 35,
}
local r8_0 = {
  [0] = 41,
  [1] = 42,
  [2] = 43,
  [3] = 44,
  [4] = 45,
  [5] = 41,
  [6] = 42,
  [7] = 43,
  [8] = 44,
  [9] = 45,
}
a24 = function()
  -- line: [105, 108] id: 3
  if a18 then
    a18:close()
  end
  if a17 then
    a17:close()
  end
end
a25 = function(r0_4)
  -- line: [111, 114] id: 4
  local data = r0_4:receive(4)
  if not data or #data ~= 4 then
    return nil
  end
  return string.unpack("<I4", data)
end
a27 = function(r0_5)
  -- line: [117, 120] id: 5
  local data = r0_5:receive(1)
  if not data or #data ~= 1 then
    return nil
  end
  local result = string.unpack("<i1", data)
  -- Validate reasonable string length
  if result < 0 or result > MAX_STRING_LENGTH then
    return nil
  end
  return result
end
a29 = function(r0_6)
  -- line: [123, 126] id: 6
  local length = a27(r0_6)
  if not length then
    return nil
  end
  return a31(r0_6, length)
end
a31 = function(r0_7, r1_7)
  -- line: [129, 132] id: 7
  if not r1_7 or r1_7 <= 0 or r1_7 > MAX_STRING_LENGTH then
    return nil
  end
  local data = r0_7:receive(r1_7)
  if not data or #data ~= r1_7 then
    return nil
  end
  return string.unpack("c" .. tostring(r1_7), data)
end
a34 = function(r0_8)
  -- line: [135, 151] id: 8
  local machine_id = a31(r0_8, 16)
  if not machine_id then
    return nil
  end
  
  local r1_8 = {
    a36 = machine_id,
  }
  
  local r2_8 = a27(r0_8)
  if not r2_8 or r2_8 > MAX_CANDIDATES_PER_BALLOT then
    return nil
  end
  
  local r3_8 = {}
  for r7_8 = 1, r2_8, 1 do
    local candidate_id = a27(r0_8)
    local candidate_name = a29(r0_8)
    local vote_count = a25(r0_8)
    local voter_name = a29(r0_8)
    local vote_type = a27(r0_8)
    
    if not candidate_id or not candidate_name or not vote_count or not voter_name or not vote_type then
      return nil
    end
    
    -- Validate vote count is reasonable
    if vote_count < 0 or vote_count > MAX_VOTE_COUNT then
      return nil
    end
    
    r3_8[r7_8] = {
      a40 = candidate_id,
      a41 = sanitize_input(candidate_name),
      a42 = vote_count,
      a43 = sanitize_input(voter_name),
      a44 = vote_type,
    }
    
    -- Check if sanitization failed
    if not r3_8[r7_8].a41 or not r3_8[r7_8].a43 then
      return nil
    end
  end
  r1_8.a38 = r3_8
  return r1_8
end
a45 = function(r0_9)
  -- line: [154, 183] id: 9
  if not r0_9 or not r0_9.a36 or not r0_9.a38 then
    return false
  end
  
  if not r6_0[r4_0.crc32(0, r0_9.a36)] then
    return false
  end
  
  -- Begin transaction for consistency
  a18:execute("BEGIN TRANSACTION;")
  
  for r4_9 = 1, #r0_9.a38, 1 do
    local r5_9 = r0_9.a38[r4_9]
    if not r5_9.a41 or not r5_9.a43 then
      a18:execute("ROLLBACK;")
      return false
    end
    
    -- Use prepared statement equivalent (parameter binding simulation)
    local stmt = "SELECT votes FROM candidates WHERE name = ?;"
    local safe_name = sanitize_input(r5_9.a41)
    if not safe_name then
      a18:execute("ROLLBACK;")
      return false
    end
    
    local r6_9 = a18:execute("SELECT votes FROM candidates WHERE name='" .. safe_name .. "';")
    if not r6_9 then
      a18:execute("ROLLBACK;")
      return false
    end
    
    local candidate_exists = r6_9:fetch()
    r6_9:close()
    
    if not candidate_exists then
      a18:execute("ROLLBACK;")
      return false
    end
    
    local safe_voter = sanitize_input(r5_9.a43)
    if not safe_voter or not r5_0[safe_voter] then
      a18:execute("ROLLBACK;")
      return false
    end
    
    if r5_9.a44 > 1 or r5_9.a44 < 0 then
      a18:execute("ROLLBACK;")
      return false
    end
  end
  
  a18:execute("COMMIT;")
  return true
end
a49 = function(r0_10)
  -- line: [186, 201] id: 10
  if not r0_10 or not r0_10.a38 then
    return false
  end
  
  -- Begin transaction
  a18:execute("BEGIN TRANSACTION;")
  
  for r4_10 = 1, #r0_10.a38, 1 do
    local r5_10 = r0_10.a38[r4_10]
    if not r5_10.a41 then
      a18:execute("ROLLBACK;")
      return false
    end
    
    local safe_name = sanitize_input(r5_10.a41)
    if not safe_name then
      a18:execute("ROLLBACK;")
      return false
    end
    
    local r6_10 = a18:execute("SELECT votes FROM candidates WHERE name='" .. safe_name .. "';")
    if not r6_10 then
      a18:execute("ROLLBACK;")
      return false
    end
    
    local r7_10 = r6_10:fetch()
    r6_10:close()
    
    if not r7_10 or type(r7_10) ~= "number" then
      a18:execute("ROLLBACK;")
      return false
    end
    
    local new_vote_count = r7_10 + r5_10.a44
    if new_vote_count < 0 or new_vote_count > MAX_VOTE_COUNT then
      a18:execute("ROLLBACK;")
      return false
    end
    
    local update_result = a18:execute("UPDATE candidates SET votes=" .. tostring(new_vote_count) .. " WHERE name='" .. safe_name .. "';")
    if not update_result then
      a18:execute("ROLLBACK;")
      return false
    end
  end
  
  a18:execute("COMMIT;")
  return true
end
a53 = function(r0_11, r1_11)
  -- line: [204, 206] id: 11
  return "\u{1b}[" .. tostring(r0_11) .. ";" .. tostring(r1_11) .. "m"
end
a56 = function()
  -- line: [209, 253] id: 12
  if not a13 then
    local r0_12 = {}
    local r1_12 = a18:execute("SELECT name, votes FROM candidates;")
    local r2_12, r3_12 = r1_12:fetch()
    while r2_12 do
      r0_12[r2_12] = r3_12
      r2_12, r3_12 = r1_12:fetch()
    end
    r1_12:close()
    print("\u{1b}[2J")
    print("\u{1b}[H")
    print(a53(30, 47) .. a14 .. a53(37, 40))
    local r4_12 = 0
    for r8_12, r9_12 in pairs(r0_12) do
      if r4_12 < r9_12 then
        r4_12 = r9_12
      end
    end
    local r5_12 = 0
    for r9_12, r10_12 in pairs(r0_12) do
      print(a53(r7_0[r5_12], 40) .. r9_12 .. " " .. tostring(r10_12) .. a53(37, 40))
      local r11_12 = ""
      if r10_12 > 0 then
        for r15_12 = 1, a8 * r10_12 / r4_12, 1 do
          r11_12 = r11_12 .. " "
        end
        print(a53(30, r8_0[r5_12]) .. r11_12 .. a53(37, 40))
      else
        print()
      end
      r5_12 = r5_12 + 1
    end
    print(a53(30, 47) .. a15 .. a53(37, 40))
  end
end

-- Improved TLS configuration
local r9_0 = {
  mode = "server",
  protocol = "tlsv1_2",
  key = "./key.pem",
  certificate = "./cert.pem",
  options = {"all", "no_sslv2", "no_sslv3", "no_tlsv1", "cipher_server_preference"},
  ciphers = "ECDHE+AESGCM:ECDHE+CHACHA20:DHE+AESGCM:DHE+CHACHA20:!aNULL:!MD5:!DSS",
  verify = "peer",
  depth = 2,
}

a64 = function()
  -- line: [265, 326] id: 13
  local success, error_msg = pcall(a16)
  if not success then
    print("Failed to initialize: " .. tostring(error_msg))
    return
  end
  
  local r0_13 = r0_0.tcp()
  if not r0_13 then
    print("Failed to create TCP socket")
    return
  end
  
  r0_13:setoption("reuseaddr", true)
  r0_13:settimeout(1)
  
  local r1_13, r2_13 = r0_13:bind("0.0.0.0", 10000) -- Bind to localhost only for security
  if not r1_13 then
    print("Bind failed: " .. tostring(r2_13))
    r0_13:close()
    return 
  end
  
  local r3_13, r4_13 = r0_13:listen(5) -- Limit connection backlog
  if not r3_13 then
    print("Listen failed: " .. tostring(r4_13))
    r0_13:close()
    return 
  end
  
  a56()
  
  while a11 do
    local r5_13, err = r0_13:accept()
    if r5_13 then
      -- Set timeout for client connections
      r5_13:settimeout(30)
      
      local ssl_conn, ssl_err = r1_0.wrap(r5_13, r9_0)
      if ssl_conn then
        local handshake_ok, handshake_err = pcall(function() ssl_conn:dohandshake() end)
        if handshake_ok then
          local r6_13 = a34(ssl_conn)
          if r6_13 then
            if a13 then
              print("Received package of votes from machine", r6_13.a36)
            end
            
            if a45(r6_13) then
              if a49(r6_13) then
                ssl_conn:send(a6)
              else
                ssl_conn:send(a7)
                if a13 then
                  print("*** Package processing failed")
                end
              end
            else
              ssl_conn:send(a7)
              if a13 then
                print("*** Package rejected - validation failed")
              end
            end
          else
            ssl_conn:send(a7)
            if a13 then
              print("*** Package rejected - malformed data")
            end
          end
        else
          if a13 then
            print("*** TLS handshake failed: " .. tostring(handshake_err))
          end
        end
        ssl_conn:close()
      else
        if a13 then
          print("*** SSL wrap failed: " .. tostring(ssl_err))
        end
        r5_13:close()
      end
      a56()
    elseif err ~= "timeout" then
      if a13 then
        print("Accept failed: " .. tostring(err))
      end
    end
  end
  
  print("\nExiting...")
  r0_13:close()
  a24()
end

a64()
