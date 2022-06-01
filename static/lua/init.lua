local init = {}

init.url = "http://localhost:3300/"
init.config = {
	ws = init.url .. "turtlews",
	turtleApi = init.url .. "api/v1/turtle/",
	worldApi = init.url .. "api/v1/world/",
	luaUrl = init.url .. "static/lua/",
	files = {
		"init.lua",
		"startup.lua",
		"skyrtle.lua",
		"websocket.lua",
	},
	debug = false,
	downloadDelay = 0.25,
	apiDelay = 5,
}

-- replace first instances of http with ws for websocket in init.config.ws
init.config.ws = init.config.ws:gsub("http", "ws")

-- check if running as sub-shell
if shell.getRunningProgram() == "rom/programs/http/wget.lua" then 
	print("[Update] Checking for updates...")
	for i, file in ipairs(init.config.files) do
		print("[Downloading]: " .. file)
		sleep(init.config.downloadDelay)
		local url = init.config.luaUrl .. file
		local file = fs.open(file, "w")
		if http.get(url).readAll() then
			file.write(http.get(url).readAll())
		else
			print("[Error]: Unable to download " .. file)
		end
		file.close()
	end
	print("[Update] Update complete")
end
if init.config.debug then
	print("[Debug] Debug mode enabled")
end

return init

-- wget run http://localhost:3300/static/lua/init.lua