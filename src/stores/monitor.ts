import { defineStore } from 'pinia'
import { getDroneStatusApi, getDronePositionApi } from '@/api/mavlink'
import type { DroneStatus, DronePosition } from '@/types/mavlink'

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
        const res = await getDroneStatusApi()
        if ((res.success || res.code === 0) && res.data) {
          this.droneStatus = res.data
          this.lastSnapshot = { type: 'status', data: res.data, timestamp: Date.now() }
        }
      } catch (error) {
        console.log('获取无人机状态失败')
        this.setError('获取无人机状态失败')
      } finally {
        this.loading = false
      }
    },
    async fetchDronePosition() {
      this.loading = true
      try {
        const res = await getDronePositionApi()
        if ((res.success || res.code === 0) && res.data) {
          this.dronePosition = res.data
          this.lastSnapshot = { type: 'position', data: res.data, timestamp: Date.now() }
        }
      } catch (error) {
        console.log('获取无人机位置失败')
        this.setError('获取无人机位置失败')
      } finally {
        this.loading = false
      }
    },
    async fetchAllData() {
      await Promise.all([
        this.fetchDroneStatus(),
        this.fetchDronePosition()
      ])

      // 模拟数据
      this.chainStatus = {
        chainId: '1',
        status: 'running',
        currentNode: 'takeoff',
        progress: 30,
        startTime: new Date().toISOString(),
        lastUpdate: new Date().toISOString()
      }

      this.systemInfo = {
        cpuUsage: 45,
        memoryUsage: 60,
        diskUsage: 30,
        networkStatus: '正常',
        systemTime: new Date().toLocaleString()
      }

      // Mock 错误数据示例
      this.errors = [
        { "id": "1", "chainId": "chain_001", "error": "起飞失败", "timestamp": "2026-04-17 10:00:00" },
        { "id": "2", "chainId": "chain_002", "error": "连接超时", "timestamp": "2026-04-17 10:05:00" }
      ]
    }
  }
})

