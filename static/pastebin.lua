local pastebin = {}
function pastebin.fetch(address)
	local cacheBuster = ("%x"):format(math.random(0, 2 ^ 30))
	local response, err = http.get(
		"https://pastebin.com/raw/" .. textutils.urlEncode(address) .. "?cb=" .. cacheBuster
	)
	if response then
		local headers = response.getResponseHeaders()
		if not headers["Content-Type"] or not headers["Content-Type"]:find("^text/plain") then
			return err, "Pastebin Spam Detection",  "https://pastebin.com/" .. textutils.urlEncode(address)
		end
		return response.readAll()
	end
	return false
end

function pastebin.get(address, name)
	if name == "" then
		name = address
	end
	local data = pastebin.fetch(address)
	return data
end

function pastebin.run(address)
	local data = pastebin.fetch(address)
	http.post(
		ultron.config.api.host.. "turtle/" .. ultron.data.id,	
		""
	)
end
return pastebin