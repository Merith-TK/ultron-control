local init = {}
local cID = os.getComputerID()
init.url = "http://localhost:3300/"
init.config = {
	debug = true,
	ws = {
		turtle = init.url .. "turtlews",
	},
	api = {
		computer = init.url .. "api/v1/computer/" .. cID.. "/",
		turtle =   init.url .. "api/v1/turtle/" .. cID.. "/",
		pocket =   init.url .. "api/v1/pocket/" .. cID.. "/",
		world  =   init.url .. "api/v1/world/" .. cID.. "/",
	},
	luaUrl = init.url .. "static/",
	files = {
		all = {
			"init.lua",
			"startup.lua",
		},
		computer = {
			"computer/Terminal.lua"
		},
		turtle = {
			"turtle/skyrtle.lua",
			"turtle/Terminal.lua",
		},
		pocket = {
			"pocket/Terminal.lua",
		},

	},
	turtle = {
		fuels = {
			"minecraft:coal",
			"minecraft:coal_block",
			"minecraft:charcoal",
		},
	},
	downloadDelay = 0.25,
	apiDelay = 5,
}

-- replace all http with ws in init.config.ws
for i, v in pairs(init.config.ws) do
	init.config.ws[i] = v:gsub("http", "ws")
end

-- download files
local function downloadFiles(files)
	for _, file in ipairs(files) do
		local url = init.config.luaUrl .. file
		local path = shell.resolve(file)
		-- download files using http
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
end

--------------------------------------------------------------------------------
-- debugPrint(string)
-- Prints a string to the terminal if debug is enabled
--------------------------------------------------------------------------------
function init.debugPrint(str)
	if init.config.debug then
		print("[Debug] "..str)
	end
end


-- check if running as sub-shell
if shell.getRunningProgram() == "rom/programs/http/wget.lua" then
	shell.run("set motd.enable false")
	term.clear()
	term.setCursorPos(1,1)

	print("[Turtle] ".. os.getComputerLabel())
	print("[Updating]: Auto-updating...")
	downloadFiles(init.config.files.all)
	if turtle then
		downloadFiles(init.config.files.turtle)
	elseif pocket then
		downloadFiles(init.config.files.pocket)
	elseif commands then
		print("[Error]: This program is not compatible with the Command Computer for security reasons.")
	else
		downloadFiles(init.config.files.computer)
	end

	print("[Update] Update complete")
	init.debugPrint("Debug mode enabled")
end

return init

-- wget run http://localhost:3300/static/init.lua