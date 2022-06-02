-- worldPost structure

--	local worldPost = {
--		[ pos.x ..",".. pos.y ..",".. pos.z ] = {	
--			block = turtle.inspect(),",
-- 			data = {},
-- 			meta = {},
-- 			misc = {},
--		},
--	}


local function postWorld(data)
	local WorldData =  textutils.serializeJSON(data)
	local file = fs.open("/WorldData.json", "w")
	file.write(WorldData)
	file.close()
end
