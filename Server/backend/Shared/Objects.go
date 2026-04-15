package SharedObjects

import (
	Drone "MavlinkProject/Server/backend/Shared/Drones"
	Device "MavlinkProject/Server/backend/Shared/Device"
	LandNode "MavlinkProject/Server/backend/Shared/LandNode"
	Charging "MavlinkProject/Server/backend/Shared/Charge"
	User "MavlinkProject/Server/backend/Shared/User"
)

var (
	UserModel         = User.User{}
	DeviceModel       = &Device.Device{}
	DroneModel        = &Drone.Drone{}
	LandNodeModel     = &LandNode.LandNode{}
	ChargingCaseModel = &Charging.ChargingCase{}
)

var ObjectModels = []interface{}{
	UserModel,
	DeviceModel,
	DroneModel,
	LandNodeModel,
	ChargingCaseModel,
}
