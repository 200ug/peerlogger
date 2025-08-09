package db

/*
	initial schema idea:

	- general: node id, ip address, ports, client version
	- metadata: last seen timestamp, geolocation data
	- status: connection status
*/

type Database struct {
	// todo: add variables necessary for storing the postgres db handler data
	// ** this is a plceholder **
}

func NewDatabase(DBURL string) *Database {
	// todo: init conn to postgres db (and make sure migrations etc. are applied properly)
	// ** this is a plceholder **
	return &Database{}
}
