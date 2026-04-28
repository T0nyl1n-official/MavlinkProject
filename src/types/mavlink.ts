import type { ApiResponse, JsonObject } from './api'

export interface DroneStatus extends JsonObject {
  armed: boolean
  mode: string
  battery: number
  altitude: number
  speed: number
  position: {
    latitude: number
    longitude: number
    altitude: number
  }
}

export interface DronePosition extends JsonObject {
  latitude: number
  longitude: number
  altitude: number
  heading: number
  speed: number
}

export interface TakeoffParams {
  system_id?: number
  altitude: number
}

export interface LandParams {
  system_id?: number
  speed?: number
}

export interface MoveParams {
  system_id?: number
  latitude: number
  longitude: number
  altitude: number
  speed?: number
}

export interface SensorAlertParams {
  sensor_id: string
  latitude: number
  longitude: number
  radius: number
  photo_count: number
  altitude: number
}

export interface CreateHandlerParams {
  handler_type: string
  config?: JsonObject
}

export interface StartConnectionParams {
  connection_type: string
  config?: JsonObject
}

export interface MavlinkConnection {
  id: string
  version: string
  ip: string
  port: number
  sysid: number
  compid: number
  connected: boolean
}

export interface ConnectParams {
  version: string
  ip: string
  port: number
  sysid: number
  compid: number
}

export interface MavlinkCommandParams {
  connectionId: string
  command: string
  params: number[]
}

export type MavlinkResponse = ApiResponse<{ message?: string }>
export type MavlinkConnectionsResponse = ApiResponse<{ connections: MavlinkConnection[] }>

export type DroneStatusResponse = ApiResponse<DroneStatus>
export type DronePositionResponse = ApiResponse<DronePosition>
export type CommandResponse = ApiResponse<{ message: string }>
export type HandlerResponse = ApiResponse<{ handler_id: string }>
export type ConnectionResponse = ApiResponse<{ status: string }>