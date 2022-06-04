local init = require("init")
shell.run("wget run " .. init.config.luaUrl .. "init.lua")

os.pullEvent=os.pullEventRaw

if turtle then
	--- check if skyrtle pos has been initialized
	--- warning, this is a hack, it will not work if the turtle is actually located at (0,0,0) and facing north
	local skyrtle = require "turtle.skyrtle"
	local x,y,z = skyrtle.getPosition()
	local facing = skyrtle.getFacing()
	if x == 0 and y == 0 and z == 0 and facing == 0 then
		-- get user input for position
		print("Please enter turtle position")
		print("Press f3 and look at turtle to get position")
		print("X Y Z Facing")
		print("Example: 0 0 0 north")
		local userPos = read()
		-- convert userpos to table seperated by spaces
		local pos = {}
		for i in string.gmatch(userPos, "%S+") do
			table.insert(pos, i)
		end
		-- create skyrtle file
		skyrtle.setPosition(tonumber(pos[1]), tonumber(pos[2]), tonumber(pos[3]))
		skyrtle.setFacing(pos[4])
	end
	print("[Skyrtle]: Location ", skyrtle.getPosition())
	print("[Skyrtle]: Facing ", skyrtle.getFacing())

	shell.run("turtle/Terminal")
elseif pocket then
	shell.run("pocket/Terminal")
elseif commands then
	print("[Error]: This program is not compatible with the Command Computer for security reasons.")
else
	shell.run("computer/Terminal")
end