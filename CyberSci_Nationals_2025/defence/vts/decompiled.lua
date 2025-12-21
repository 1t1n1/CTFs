-- Decompiled using luadec 2.2 rev: 158e15a for Lua 5.3 from https://github.com/zhangjiequan/luadec
-- Command line: ./vts.luc 

-- params : ...
-- function num : 0 , upvalues : _ENV
local socket = require("socket")
local ssl = require("ssl")
local luasql_sqlite3 = require("luasql.sqlite3")
local posix_signal = require("posix.signal")
local crc32 = require("crc32")
return_1 = "\001"
return_nul = "\000"
num_100 = 100
local voter_names = {}
local crc_dict = {}
is_running = true
a12 = function(num)
  -- function num : 0_0 , upvalues : _ENV
  is_running = false
end

verbose = false
conn_message_1 = "              *** Val Verde Central Electoral Commission * Vote Tabulation Service ***              "
conn_message_2 = " Press Ctrl+C to exit                                                                               "
init = function()
  -- function num : 0_1 , upvalues : _ENV, posix_signal, luasql_sqlite3, voter_names, crc_dict
  if arg[1] == "-s" then
    verbose = true
    print(conn_message_1)
    print(conn_message_2)
  end
  ;
  (posix_signal.signal)(posix_signal.SIGINT, a12)
  a17 = (luasql_sqlite3.sqlite3)()
  a18 = a17:connect("./vts.db")
  local a19 = a18:execute("SELECT name FROM voters;")
  local name = a19:fetch()
  while name do
    voter_names[name] = true
    name = a19:fetch()
  end
  a19:close()
  local a21 = nil
  a19 = a18:execute("SELECT crc FROM machines;")
  a21 = a19:fetch()
  while a21 do
    crc_dict[a21] = true
    a21 = a19:fetch()
  end
  a19:close()
end

-- Question: What are they used for??
local dict_1 = {[0] = 31, [1] = 32, [2] = 33, [3] = 34, [4] = 35, [5] = 31, [6] = 32, [7] = 33, [8] = 34, [9] = 35}
local dict_2 = {[0] = 41, [1] = 42, [2] = 43, [3] = 44, [4] = 45, [5] = 41, [6] = 42, [7] = 43, [8] = 44, [9] = 45}
close_sql_req = function()
  -- function num : 0_2 , upvalues : _ENV
  a18:close()
  a17:close()
end

extract_unsigned_int = function(c)
  -- function num : 0_3 , upvalues : _ENV
  local a26 = c:receive(4)
  return (string.unpack)("<I4", a26)
end

extract_signed_byte = function(c)
  -- function num : 0_4 , upvalues : _ENV
  local a28 = c:receive(1)
  return (string.unpack)("<i1", a28)
end

extract_var_size_string = function(c)
  -- function num : 0_5 , upvalues : _ENV
  local a30 = extract_signed_byte(c)
  return extract_string(c, a30)
end

extract_string = function(c, size)
  -- function num : 0_6 , upvalues : _ENV
  local a32 = c:receive(size)
  return (string.unpack)("c" .. tostring(size), a32)
end

extract_data_from_req = function(conn)
  -- function num : 0_7 , upvalues : _ENV
  local data = {}
  data.string1 = extract_string(conn, 16)
  local n_elems = extract_signed_byte(conn)
  local inner_dict = {}
  for i = 1, n_elems do
    local a39 = {}
    a39.a40 = extract_signed_byte(conn)
    a39.candidate_vote = extract_var_size_string(conn)
    a39.a42 = extract_unsigned_int(conn)
    a39.voter_name = extract_var_size_string(conn)
    a39.prob_number_of_votes = extract_signed_byte(conn)
    inner_dict[i] = a39
  end
  data.inner_dict = inner_dict
  if verbose then
    print("data.string1: " .. data.string1)
    for i = 1, n_elems do
      print("data.inner_dict["..i.."]:")
      for key,value in pairs(data.inner_dict[i]) do
        print("Key: " .. tostring(key) .. ", Value: " .. tostring(value))
      end
    end
    print()
  end
  return data
end

allowed_vote = function(data)
  -- function num : 0_8 , upvalues : crc_dict, crc32, _ENV, voter_names
  -- Prob qqch ici
  if not crc_dict[(crc32.crc32)(0, data.string1)] then
    return false
  end
  for i = 1, #data.inner_dict do
    local a47 = (data.inner_dict)[i]
    local a19 = a18:execute("SELECT votes FROM candidates WHERE name=\'" .. a47.candidate_vote .. "\';")
    local a48 = a19:fetch()
    if not a48 then
      a19:close()
      return false
    end
    if not voter_names[a47.voter_name] then
      return false
    end
    if a47.prob_number_of_votes > 1 or a47.prob_number_of_votes < 0 then
      return false
    end
  end
  return true
end

apply_vote = function(a50)
  -- function num : 0_9 , upvalues : _ENV
  for i = 1, #a50.inner_dict do
    local a51 = (a50.inner_dict)[i]
    local a19 = a18:execute("SELECT votes FROM candidates WHERE name=\'" .. a51.candidate_vote .. "\';")
    local a52 = a19:fetch()
    a19:close()
    a52 = a52 + a51.prob_number_of_votes
    a19 = a18:execute("UPDATE candidates SET votes=" .. tostring(a52) .. " WHERE name=\'" .. a51.candidate_vote .. "\';")
  end
end

fancy_print = function(a54, a55)
  -- function num : 0_10 , upvalues : _ENV
  return "\027[" .. tostring(a54) .. ";" .. tostring(a55) .. "m"
end

print_status = function()
  -- function num : 0_11 , upvalues : _ENV, dict_1, dict_2
  if not verbose then
    local name_votes_dict = {}
    local a19 = a18:execute("SELECT name, votes FROM candidates;")
    local name, votes = a19:fetch()
    while name do
      name_votes_dict[name] = votes
      name = a19:fetch()
    end
    a19:close()
    print("\027[2J")
    print("\027[H")
    print(fancy_print(30, 47) .. conn_message_1 .. fancy_print(37, 40))
    local max_number_of_votes = 0
    for _,votes in pairs(name_votes_dict) do
      if max_number_of_votes < votes then
        max_number_of_votes = votes
      end
    end
    local idx = 0
    for name,votes in pairs(name_votes_dict) do
      print(fancy_print(dict_1[idx], 40) .. name .. " " .. tostring(votes) .. fancy_print(37, 40))
      local total_string = ""
      if votes > 0 then
        for i = 1, num_100 * (votes / max_number_of_votes) do
          total_string = total_string .. " "
        end
        print(fancy_print(30, dict_2[idx]) .. total_string .. fancy_print(37, 40))
      else
        print()
      end
      idx = idx + 1
    end
    print(fancy_print(30, 47) .. conn_message_2 .. fancy_print(37, 40))
  end
end

local params = {mode = "server", protocol = "tlsv1_2", key = "./key.pem", certificate = "./cert.pem", options = "all"}
main = function()
  -- function num : 0_12 , upvalues : _ENV, socket, ssl, params
  init()
  local server = (socket.tcp)()
  server:setoption("reuseaddr", true)
  server:settimeout(1)
  local a66, a67 = server:bind("0.0.0.0", 10000)
  if not a66 then
    print(a67)
    return 
  end
  local a68, a69 = server:listen()
  if not a68 then
    print(a69)
    return 
  end
  print_status()
  while is_running do
    local conn = server:accept()
    if conn then
      conn = (ssl.wrap)(conn, params)
      if conn then
        conn:dohandshake()
        local data = extract_data_from_req(conn)
        if verbose then
          print("Received package of votes from machine", data.string1)
        end
        local a72 = allowed_vote(data)
        if a72 then
          apply_vote(data)
          conn:send(return_1)
        else
          conn:send(return_nul)
          if verbose then
            print("*** Package rejected")
          end
        end
        conn:close()
        print_status()
      end
    end
  end
  do
    print("\nExiting...")
    server:close()
    close_sql_req()
  end
end

main()

