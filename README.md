
# API Documentation
## Backend
* `/api/v1/turtle`
    * GET: return all turtle data
* `/api/v1/turtle/<id>`
    * GET: return data for turtle with that ID
    * POST: update data for turtle with ID
        * if ID is invalid, make new turtle entry in data store
        * TODO: Restrict access to this method to specific address for security (default to localhost)
    * `/inventory`
        * GET: return inventory of turtle
    * `/pos`
        * GET: return X, Y, Z, and Rotation
    * `/name`
        * GET: return turtle name
    * `/selectedSlot`
        * GET: current selected slot in inventory