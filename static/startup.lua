_G.ultron = require("ultron")
term.clear()
term.setCursorPos(1, 1)

if not turtle then print("This program can only run on a Turtle currently") end

ultron.config.api.ws = ultron.config.api.host .. "/" .. "turtle" .. "/ws"

if not fs.exists("cfg/disableUpdate") then
    ultron.wget("startup.lua", ultron.config.api.host .. "/static/startup.lua")
    ultron.wget("ultron.lua", ultron.config.api.host .. "/static/ultron.lua")
    ultron.wget("turtle.lua", ultron.config.api.host .. "/static/turtle.lua")
	ultron.wget("pastebin.lua", ultron.config.api.host .. "/static/pastebin.lua")
else
    ultron.debugPrint("Update is disabled")
end

shell.run("turtle.lua")
