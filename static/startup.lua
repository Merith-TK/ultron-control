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

-- Load agent helpers into global env before turtle.lua starts its main loop.
-- skyrtle is not available yet, but agent.lua only references it inside
-- function bodies, so it is looked up at call-time after turtle.lua loads it.
if fs.exists("cfg/agenttools") then
    ultron.wget("agent.lua", ultron.config.api.host .. "/static/agent.lua")
    _G.agentmcp = require("agent")
    ultron.debugPrint("Agent tools loaded")
end

shell.run("turtle.lua")
