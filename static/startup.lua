local ultron = require("ultron")
term.clear()
term.setCursorPos(1,1)

local currentModule = ""

if not fs.exists("cfg/module") then
	print("Which module do you want to use?")
	for i, module in ipairs(ultron.modules) do
		print(i .. ": " .. module.Name)
	end
	local choice = tonumber(read())
	-- check if choice is valid
	if choice and choice > 0 and choice <= #ultron.modules then
		local module = ultron.modules[choice]
		ultron.download_module(module.Name)
	end
	if not fs.exists("cfg") then
		fs.makeDir("cfg")
	end
	local cfg = fs.open("cfg/module", "w")
	cfg.write(ultron.modules[choice].Name)
	cfg.close()
else
	local cfg = fs.open("cfg/module", "r")
	currentModule = cfg.readAll()
	ultron.module = currentModule
	ultron.config.api.ws = ultron.config.api.host .. "/" .. currentModule .. "/ws"
	ultron.download_module(currentModule)
	cfg.close()
end

ultron.wget("startup.lua", ultron.config.api.host .. "/static/startup.lua")
ultron.wget("ultron.lua", ultron.config.api.host .. "/static/ultron.lua")
ultron.wget("module.lua", ultron.config.api.host .. "/"..currentModule.."/fs/module.lua")

shell.run("module.lua")