# Ultron Control
This is intended to be an api server for remotely controlling turtles, 
in its current state it is just an backend able to recieve data from an turtle, store it, and then present the data

### ToDo
	- [ ] create turtle control api (reliant on previous)
	- [ ] make frontend
	


# API Documentation
## Backend
* `/api/v1/turtle`
    * GET: return all turtle data
* `/api/v1/turtle/<id>`
    * GET: return data for turtle with that ID
    * POST: `["<lua code>"]` add code to command que to run on turtle,
		* `/inventory`
			* GET: return inventory of turtle
		* `/pos`
			* GET: return X, Y, Z, and Rotation
		* `/name`
			* GET: return turtle name
		* `/selectedSlot`
			* GET: current selected slot in inventory