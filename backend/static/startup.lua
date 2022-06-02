local init = require("init")
shell.run("wget run " .. init.config.luaUrl .. "init.lua")

os.pullEvent=os.pullEventRaw

if turtle then
	shell.run("turtle/Terminal")
elseif pocket then
	shell.run("pocket/Terminal")
elseif commands then
	print("[Error]: This program is not compatible with the Command Computer for security reasons.")
else
	shell.run("computer/Terminal")
end