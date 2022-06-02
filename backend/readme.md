to set up an turtle, please allow `localhost` in your computercraft server config (`saves/<worldname>/serverconfig/computercraft-server.toml`)

and then `wget run http://localhost:3300/static/lua/init.lua`

to adjust config for all turtles accessing this system, modify `init.config` table in `init.lua` directly in the file `/static/lua/init.lua` and then reboot the computer/turtle

turtle uses skyrtle for movement

## deb shit

test cmd api with
`curl -X POST -H "Content-Type: application/json"  --data-binary '["return turtle.turnLeft()"]' http://localhost:3300/api/v1/turtle/1`