import { defineStore } from 'pinia'
import { getDroneStatusApi, getDronePositionApi } from '@/api/mavlink'
import type { DroneStatus, DronePosition } from '@/types/mavlink'
import { USE_REAL_API } from '@/utils/constants'

interface ChainStatus {
  chainId: string
  status: string
  currentNode: string
  progress: number
  startTime: string
  lastUpdate: string
}

interface SystemInfo {
  cpuUsage: number
  memoryUsage: number
  diskUsage: number
  networkStatus: string
  systemTime: string
}

interface ErrorInfo {
  id: string
  chainId: string
  error: string
  timestamp: string
}

interface MonitorState {
  lastError: string | null
  lastSnapshot: Record<string, unknown> | null
  droneStatus: DroneStatus | null
  dronePosition: DronePosition | null
  chainStatus: ChainStatus | null
  systemInfo: SystemInfo | null
  errors: ErrorInfo[]
  loading: boolean
}

function createSnapshotRecord(kind: string, payload: unknown) {
  return { type: kind, data: payload, timestamp: Date.now() }
}

function createMockDroneStatus(): DroneStatus {
  return {
    armed: false,
    mode: 'STABILIZE',
    battery: 85,
    altitude: 10,
    speed: 0,
    position: {
      latitude: 22.5431,
      longitude: 114.0523,
      altitude: 10
    }
  }
}

function createMockDronePosition(): DronePosition {
  return {
    latitude: 22.5431,
    longitude: 114.0523,
    altitude: 10,
    heading: 45,
    speed: 0
  }
}

function createMockChainStatus(): ChainStatus {
  return {
    chainId: '1',
    status: 'running',
    currentNode: 'takeoff',
    progress: 30,
    startTime: new Date().toISOString(),
    lastUpdate: new Date().toISOString()
  }
}

function createMockSystemInfo(): SystemInfo {
  return {
    cpuUsage: 45,
    memoryUsage: 60,
    diskUsage: 30,
    networkStatus: '正常',
    systemTime: new Date().toLocaleString()
  }
}

function createMockErrors(): ErrorInfo[] {
  return [
    { id: '1', chainId: 'chain_001', error: '起飞失败', timestamp: '2026-04-17 10:00:00' },
    { id: '2', chainId: 'chain_002', error: '连接超时', timestamp: '2026-04-17 10:05:00' }
  ]
}

export const useMonitorStore = defineStore('monitor', {
  state: (): MonitorState => ({
    lastError: null,
    lastSnapshot: null,
    droneStatus: null,
    dronePosition: null,
    chainStatus: null,
    systemInfo: null,
    errors: [],
    loading: false
  }),
  actions: {
    setSnapshot(snapshot: Record<string, unknown>) {
      this.lastSnapshot = snapshot
      this.lastError = null
    },
    setError(message: string) {
      this.lastError = message
    },
    async fetchDroneStatus() {
      this.loading = true
      try {
        if (USE_REAL_API) {
          const response = await getDroneStatusApi()
          if ((response.success || response.code === 0) && response.data) {
            this.droneStatus = response.data
            this.lastSnapshot = createSnapshotRecord('status', response.data)
          }
        } else {
          this.droneStatus = createMockDroneStatus()
          this.lastSnapshot = createSnapshotRecord('status', this.droneStatus)
        }
      } catch (error) {
        console.error('[monitor] 获取无人机状态失败', error)
        this.setError('无人机状态暂时拿不到')
      } finally {
        this.loading = false
      }
    },
    async fetchDronePosition() {
      this.loading = true
      try {
        if (USE_REAL_API) {
          const response = await getDronePositionApi()
          if ((response.success || response.code === 0) && response.data) {
            this.dronePosition = response.data
            this.lastSnapshot = createSnapshotRecord('position', response.data)
          }
        } else {
          this.dronePosition = createMockDronePosition()
          this.lastSnapshot = createSnapshotRecord('position', this.dronePosition)
        }
      } catch (error) {
        console.error('[monitor] 获取无人机位置失败', error)
        this.setError('无人机位置暂时拿不到')
      } finally {
        this.loading = false
      }
    },
    async fetchAllData() {
      await Promise.all([
        this.fetchDroneStatus(),
        this.fetchDronePosition()
      ])

      this.chainStatus = createMockChainStatus()
      this.systemInfo = createMockSystemInfo()
      this.errors = createMockErrors()
    }
  }
})

