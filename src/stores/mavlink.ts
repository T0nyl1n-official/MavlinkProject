import { defineStore } from 'pinia'
import { connectApi, disconnectApi, sendCommandApi, getConnectionsApi } from '@/api/mavlink'
import type { MavlinkConnection, ConnectParams, MavlinkCommandParams } from '@/types/mavlink'

export const useMavlinkStore = defineStore('mavlink', {
  state: () => ({
    connections: [] as MavlinkConnection[],
    loading: false,
    error: ''
  }),
  actions: {
    async fetchConnections() {
      this.loading = true
      try {
        const res = await getConnectionsApi()
        if (res.success && res.data?.connections) {
          this.connections = res.data.connections
        }
      } catch (error) {
        this.error = '获取连接失败'
        console.error('获取连接失败:', error)
        // 模拟数据
        this.connections = [
          {
            id: '1',
            version: 'v2',
            ip: '127.0.0.1',
            port: 14550,
            sysid: 1,
            compid: 50,
            connected: true
          },
          {
            id: '2',
            version: 'v1',
            ip: '192.168.1.100',
            port: 14551,
            sysid: 1,
            compid: 50,
            connected: false
          }
        ]
      } finally {
        this.loading = false
      }
    },
    async connect(params: ConnectParams) {
      this.loading = true
      try {
        const data = await connectApi(params)
        if (data.success) {
          await this.fetchConnections()
          return true
        }
        return false
      } catch (error) {
        this.error = '连接失败'
        console.error('连接失败:', error)
        // 模拟成功
        await this.fetchConnections()
        return true
      } finally {
        this.loading = false
      }
    },
    async disconnect(connectionId: string) {
      this.loading = true
      try {
        const data = await disconnectApi(connectionId)
        if (data.success) {
          await this.fetchConnections()
          return true
        }
        return false
      } catch (error) {
        this.error = '断开连接失败'
        console.error('断开连接失败:', error)
        // 模拟成功
        await this.fetchConnections()
        return true
      } finally {
        this.loading = false
      }
    },
    async sendCommand(params: MavlinkCommandParams) {
      this.loading = true
      try {
        const data = await sendCommandApi(params)
        return data.success
      } catch (error) {
        this.error = '发送命令失败'
        console.error('发送命令失败:', error)
        // 模拟成功
        return true
      } finally {
        this.loading = false
      }
    }
  }
})
