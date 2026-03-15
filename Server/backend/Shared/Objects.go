package SharedObjects

import (
	Drone "MavlinkProject/Server/backend/Shared/Drones"
	LandNode "MavlinkProject/Server/backend/Shared/LandNode"
	Charging "MavlinkProject/Server/backend/Shared/Charge"
	User "MavlinkProject/Server/backend/Shared/User"
)

var (
	UserModel         = User.User{}
	DroneModel        = &Drone.Drone{}
	LandNodeModel     = &LandNode.LandNode{}
	ChargingCaseModel = &Charging.ChargingCase{}
)

var ObjectModels = []interface{}{
	UserModel,
	DroneModel,
	LandNodeModel,
	ChargingCaseModel,
}
