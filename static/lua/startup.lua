
local init = require("init")
shell.run("wget run " .. init.config.luaUrl .. "init.lua")
shell.run("websocket")