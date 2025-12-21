local socket = require("socket")
local ssl = require("ssl")
local luasql_sqlite3 = require("luasql.sqlite3")
local posix_signal = require("posix.signal")
local crc32 = require("crc32")

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


if not crc_dict[(crc32.crc32)(0, data.string1)] then
	return false
end
