package SharedObjects

import (
	Drone "MavlinkProject/Server/backend/Shared/Drones"
	LandNode "MavlinkProject/Server/backend/Shared/LandNode"
	User "MavlinkProject/Server/backend/Shared/User"
)

var (
	UserModel     = User.User{}
	DroneModel    = &Drone.Drone{}
	LandNodeModel = &LandNode.LandNode{}
)

var ObjectModels = []interface{}{
	UserModel,
	DroneModel,
	LandNodeModel,
}
