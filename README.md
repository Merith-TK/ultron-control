# Ultron Control
This is intended to be an api server for remotely controlling turtles, 
in its current state it is just an backend able to recieve data from an turtle, store it, and then present the data

### ToDo
	- [ ] create turtle control api (reliant on previous)
	- [ ] make frontend
	
### Install on Turtle
1. configure `ultron.lua` with your host address (this is the address that the turtle will talk to)
2. place the turtle facing North and then run `wget run <address>/static/ultron.lua` on the turtle
4. reboot turtle
5. Terminate program (hold ctrl+t for a few seconds or click the button)
6. open f3, and then open `lua` on the turtle,
7. do `skyrtle.setPosition(x,y,z)` using the "targeted block" on the right hand side
8. Reboot turtle
9. Optional: if you want to turn on the "sight" of the turtle, create the file `cfg/inspectWorld` on the turtle, (if you use the edit command to do it, remove the .lua at the end of the file name)

# API Documentation
## Backend
* `/api/turtles`
    * GET: return all turtle data
* `/api/turtle/<id>`
    * GET: return data for turtle with that ID
	* `/inventory`
			* GET: return inventory of turtle
		* `/pos`
			* GET: return X, Y, Z, and Rotation
		* `/name`
			* GET: return turtle name
		* `/selectedSlot`
			* GET: current selected slot in inventory
    * POST: `["<lua code with \escaped \"quotes\">"]` add code to command que to run on turtle,
