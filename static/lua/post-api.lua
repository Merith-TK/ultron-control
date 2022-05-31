----
-- this file contains the basic api for the
-- Ultron Turtle Control API
----

-- TODO: add a way to get the current rotation
-- TODO:  get world api operational
-- TODO: fix turtle inspect front

--assert(turtle, "This must be ran on an TURTLE")
assert(os.getComputerID() ~= 0, "This cannot be ran on ID 0")

local skyrtle = require("skyrtle")
skyrtle.hijack(true)
local turtlePost = {
	name = "",
	id = 0,
	pos = {
		x = 0,
		y = 0,
		z = 0,
		r = 0
	},
	fuel = 0,
	maxFuel = 0,
	selectedSlot = 0,
	inventory = {
	},
}

local worldPost = {
	block = "",
	pos = {
		x = 0,
		y = 0,
		z = 0,
	},
	data = {},
}


local function postWorld(data)
local WorldData =  textutils.serializeJSON(data)
local file = fs.open("/WorldData.json", "w")
file.write(WorldData)
file.close()
end
local function postTurtle()
	local TurtleData =  textutils.serializeJSON(turtlePost)
	http.post("http://localhost:3300/api/v1/turtle/"..tostring(turtlePost.id), TurtleData)
end
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
	local data = {
		x = turtlePost.pos.x,
		y = turtlePost.pos.y,
		z = turtlePost.pos.z,
		r = turtlePost.pos.r,
	}
	-- calculate the position of the block infront of the turtle
	local r, rname = skyrtle.getFacing()
	if rname == "north" then
		data.z = data.z - 1
	elseif rname == "south" then
		data.z = data.z + 1
	elseif rname == "east" then
		data.z = data.x + 1
	elseif rname == "west" then
		data.z = data.x - 1
	end


	-- scan for blocks
	local front, frontData = turtle.inspect()
	local up, upData = turtle.inspectUp()
	local down, downData = turtle.inspectDown()
	-- if there is a block infront, add it to the worldPost
	if front then
		local post = worldPost
		post.data = frontData
		post.block = frontData.name
		-- TODO: add a way to get the postion of the block infront of the turtle
		--postWorld(post)
		print("ERR: unable to post FrontData due to positioning being borked")
	else
		local post = worldPost
		post.data = {}
		post.block = ""
		postWorld(post)
	end
	-- if there is a block above, add it to the worldPost
	if up then
		local post = worldPost
		post.data = upData
		post.block = upData.name
		post.pos = {
			x = turtlePost.pos.x,
			y = turtlePost.pos.y + 1,
			z = turtlePost.pos.z,
		}
		postWorld(post)
	else
		local post = worldPost
		post.data = {}
		post.block = ""
		post.pos = {
			x = turtlePost.pos.x,
			y = turtlePost.pos.y + 1,
			z = turtlePost.pos.z,
		}
		postWorld(post)
	end
	-- if there is a block below, add it to the worldPost
	if down then
		local post = worldPost
		post.data = downData
		post.block = downData.name
		post.pos = {
			x = turtlePost.pos.x,
			y = turtlePost.pos.y - 1,
			z = turtlePost.pos.z,
		}
		postWorld(post)
	else
		local post = worldPost
		post.data = {}
		post.block = ""
		post.pos = {
			x = turtlePost.pos.x,
			y = turtlePost.pos.y - 1,
			z = turtlePost.pos.z,
		}
		postWorld(post)
	end
end


local function postData()
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

	turtlePost.fuel = turtle.getFuelLevel()
	turtlePost.maxFuel = turtle.getFuelLimit()

	turtlePost.selectedSlot = turtle.getSelectedSlot()
	getInventory()

	postTurtle()
	inspectWorld()
end

postData()
