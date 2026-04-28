import request from './request'
import { mockService, config } from '@/utils/mockService'
import { ElMessage } from 'element-plus'
import type {
  CommandResponse,
  ConnectionResponse,
  ConnectParams,
  CreateHandlerParams,
  DronePositionResponse,
  DroneStatusResponse,
  HandlerResponse,
  LandParams,
  MoveParams,
  MavlinkCommandParams,
  MavlinkConnectionsResponse,
  MavlinkResponse,
  SensorAlertParams,
  StartConnectionParams,
  TakeoffParams
} from '@/types/mavlink'

const showMockMessage = () => {
  ElMessage.info('当前为 Mock 模式，联调时替换为真实接口')
}

export function createHandlerApi(params: CreateHandlerParams): Promise<HandlerResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      data: {
        handler_id: `handler-${Date.now()}`
      },
      message: 'Handler created successfully'
    })
  }
  return request.post('/mavlink/v1/handler/create', params)
}

export function deleteHandlerApi(id: string): Promise<CommandResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Handler deleted successfully'
    })
  }
  return request.delete(`/mavlink/v1/handler/${id}`)
}

export function startConnectionApi(params: StartConnectionParams): Promise<ConnectionResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      data: {
        status: 'connected'
      },
      message: 'Connection started successfully'
    })
  }
  return request.post('/mavlink/v1/connection/start', params)
}

export function stopConnectionApi(): Promise<ConnectionResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      data: {
        status: 'disconnected'
      },
      message: 'Connection stopped successfully'
    })
  }
  return request.post('/mavlink/v1/connection/stop')
}

export function takeoffApi(params: TakeoffParams): Promise<CommandResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Takeoff command sent successfully'
    })
  }
  return request.post('/mavlink/v1/drone/takeoff', params)
}

export function landApi(params?: LandParams): Promise<CommandResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Land command sent successfully'
    })
  }
  return request.post('/mavlink/v1/drone/land', params || {})
}

export function moveApi(params: MoveParams): Promise<CommandResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Move command sent successfully'
    })
  }
  return request.post('/mavlink/v1/drone/move', params)
}

export function returnApi(): Promise<CommandResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Return command sent successfully'
    })
  }
  return request.post('/mavlink/v1/drone/return')
}

export function getDroneStatusApi(): Promise<DroneStatusResponse> {
  if (!config.USE_REAL_API) {
    return Promise.resolve({
      code: 0,
      success: true,
      data: {
        armed: false,
        mode: 'STABILIZE',
        battery: 85,
        altitude: 0,
        speed: 0,
        position: {
          latitude: 31.2304,
          longitude: 121.4737,
          altitude: 0
        }
      },
      message: 'Drone status retrieved successfully'
    })
  }
  return request.get('/mavlink/v1/drone/status')
}

export function getDronePositionApi(): Promise<DronePositionResponse> {
  if (!config.USE_REAL_API) {
    return Promise.resolve({
      code: 0,
      success: true,
      data: {
        latitude: 31.2304,
        longitude: 121.4737,
        altitude: 0,
        heading: 0,
        speed: 0
      },
      message: 'Drone position retrieved successfully'
    })
  }
  return request.get('/mavlink/v1/drone/position')
}

export function v2TakeoffApi(params: TakeoffParams): Promise<CommandResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Takeoff command sent successfully'
    })
  }
  return request.post('/mavlink/v2/takeoff', params)
}

export function v2LandApi(params?: LandParams): Promise<CommandResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Land command sent successfully'
    })
  }
  return request.post('/mavlink/v2/land', params || {})
}

export function v2MoveApi(params: MoveParams): Promise<CommandResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Move command sent successfully'
    })
  }
  return request.post('/mavlink/v2/move', params)
}

export function sensorAlertApi(params: SensorAlertParams): Promise<CommandResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Sensor alert sent successfully'
    })
  }
  return request.post('/api/sensor/message', {
    sensor_id: params.sensor_id,
    sensor_ip: '127.0.0.1',
    sensor_name: params.sensor_id,
    alert_type: 'fire',
    alert_msg: '前端模拟告警',
    latitude: params.latitude,
    longitude: params.longitude,
    timestamp: Math.floor(Date.now() / 1000),
    severity: 'high'
  })
}

export function connectApi(params: ConnectParams): Promise<MavlinkResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return mockService.connectMavlink('mavlink-001').then(() => ({
      code: 0,
      success: true,
      message: 'Connected successfully'
    }))
  }
  return request.post('/mavlink/connect', params)
}

export function disconnectApi(connectionId: string): Promise<MavlinkResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return mockService.disconnectMavlink(connectionId).then(() => ({
      code: 0,
      success: true,
      message: 'Disconnected successfully'
    }))
  }
  return request.post('/mavlink/disconnect', { connectionId })
}

export function sendCommandApi(params: MavlinkCommandParams): Promise<MavlinkResponse> {
  if (!config.USE_REAL_API) {
    showMockMessage()
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Command sent successfully'
    })
  }
  return request.post('/mavlink/send-command', params)
}

export function getConnectionsApi(): Promise<MavlinkConnectionsResponse> {
  if (!config.USE_REAL_API) {
    return mockService.getMavlinkDevices().then(devices => ({
      code: 0,
      success: true,
      data: {
        connections: devices
      },
      message: 'Connections retrieved successfully'
    }))
  }
  return request.get('/api/sensor/status').then((res: any) => ({
    code: typeof res?.code === 'number' ? res.code : 0,
    success: res?.success ?? res?.code === 0,
    message: res?.message || 'Connections retrieved successfully',
    data: {
      connections: Array.isArray(res?.drones)
        ? res.drones.map((drone: any) => ({
            id: drone.board_id || drone.system_id || '',
            version: 'v1',
            ip: drone.ip || '',
            port: Number(drone.port || 0),
            sysid: Number(drone.system_id || 0),
            compid: Number(drone.compid || 0),
            connected: Boolean(drone.is_idle ?? true)
          }))
        : []
    }
  }))
}