# ultron api

- [/api] Generic API
	- [/turtle] Turtle Specific Api tree
	* GET: Fetch all turtle data
		- [/{id}]
		* GET: Fetch Turtle Data
			- [/pos]
			* POST: update database with turtpe position
			* GET: return turtle position
			* `{"x": 0, "y": 0, "z": 0, "r": [ "1/2/3/4", "north/east/south/west"]}`

## WIP
	- [/world]
		- [/block]
		* POST: `{"block":"minecraft:stone","x": 0, "y": 0, "data":[<parsed turtle.inspect output>]}`


		- [/chest]
			- POST: `{"size":27, storage:[{id:"minecraft:stick"stack":<size of stack>}]}`