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