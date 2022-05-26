# ultron api

- [/api] Generic API
    - [/turtle] Turtle Specific Api tree
        - [/pos]
            * POST: V V V
            * GET: return array of turtle positions
				```json
				[
					"id": {
						"x": 0,
						"y": 0,
						"z": 0,
						"r": [ 
							1, "north"
						]
						//1     2    3     4", 
						//north east south west"
					}
				]
				```
        - [/pos/<id#>]
            * GET: `{"x": 0, "y": 0, "z": 0, "r": [ "1/2/3/4", "north/east/south/west"]}`

    - [/world]
        - [/block]
            - POST: `{"block":"minecraft:stone","x": 0, "y": 0, "data":[<parsed turtle.inspect output>]}`
            - GET:
                - IN: x,y,z
                - OUT: block:id, data
        - [/chest]
            - POST: `{"size":27, storage:[{id:"minecraft:stick"stack":<size of stack>}]}`