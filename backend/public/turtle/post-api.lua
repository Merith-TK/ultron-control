assert(turtle, "This must be ran on an TURTLE")
assert(os.getComputerID() ~= 0, "This cannot be ran on ID 0")

local skyrtle = require("skyrtle")
local turtlePost = {
	name = "",
	id = 0,
	pos = {
		x = 0,
		y = 0,
		z = 0,
		r = 0
	},
	selectedSlot = 0,
	inventory = {
	},
}

local worldPost = {
	{
		name = "",
		pos = {
			x = 0,
			y = 0,
			z = 0,
		},
		data = {},
	}
}


local function getInventory()
	for i = 1, 16 do
		turtle.select(i)
		local item = turtle.getItemDetail(i, true)
		if not item then
			item = {
				name = "",
				count = 0,
				slot = i,
			}
		end
		turtlePost.inventory[i] = item
	end
	turtle.select(turtlePost.selectedSlot)
end

-- function to inspect the block above, below, and infront
local function inspectWorld()
	-- inspect up down and front
	local up, upData = turtle.inspectUp()
	local down, downData = turtle.inspectDown()
	local front, frontData = turtle.inspect()

	-- if there is a block above, add it to the worldPost
	if up then
		local x,y,z = skyrtle.getPosition()
		table.insert(worldPost, {
			name = upData.name,
			pos = {
				x = x,
				y = y + 1,
				z = z,
			},
			data = upData,
		})
	end
	-- if there is a block below, add it to the worldPost
	if down then
		local x,y,z = skyrtle.getPosition()
		table.insert(worldPost, {
			name = downData.name,
			pos = {
				x = x,
				y = y - 1,
				z = z,
			},
			data = downData,
		})
	end
	-- if there is a block infront, add it to the worldPost
	if front then
		-- calculate the position of the block infront using turtle rotation
		local x,y,z = skyrtle.getPosition()
		if skyrtle.getFacing() == "north" then
			z = z + 1
		elseif skyrtle.getFacing() == "east" then
			x = x + 1
		elseif skyrtle.getFacing() == "south" then
			z = z - 1
		elseif skyrtle.getFacing() == "west" then
			x = x - 1
		end
		table.insert(worldPost, {
			name = frontData.name,
			pos = {
				x = x,
				y = y,
				z = z,
			},
			data = frontData,
		})
	end
end


local function getData()

	turtlePost.id = os.getComputerID()

	local label = os.getComputerLabel()
	if label and not label == "" then
		turtlePost.name = label
	else
		os.setComputerLabel(tostring(turtlePost.id))
		turtlePost.name = tostring(turtlePost.id)
	end

	local x,y,z = skyrtle.getPosition()
	local r = skyrtle.getFacing()
	turtlePost.pos.x = x
	turtlePost.pos.y = y
	turtlePost.pos.z = z
	turtlePost.pos.r = r
	turtlePost.selectedSlot = turtle.getSelectedSlot()
	getInventory()


	inspectWorld()
end

getData()
local TurtleData =  textutils.serializeJSON(turtlePost)
local file = fs.open("/TurtleData.json", "w")
file.write(TurtleData)
file.close()

local _, inspect = turtle.inspect()
local WorldData =  textutils.serialize(worldPost)
local file = fs.open("/WorldData.lua", "w")
file.write("local data = " .. WorldData)
file.close()



http.post("http://localhost:3300/api/v1/turtle/"..tostring(turtlePost.id), TurtleData)
