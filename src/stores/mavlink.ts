import { defineStore } from 'pinia'
import { connectApi, disconnectApi, sendCommandApi, getConnectionsApi } from '@/api/mavlink'
import type { MavlinkConnection, ConnectParams, MavlinkCommandParams } from '@/types/mavlink'

function getFallbackConnections(): MavlinkConnection[] {
  return [
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
}

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
        const response = await getConnectionsApi()
        if (response.success && response.data?.connections) {
          this.connections = response.data.connections
          this.error = ''
        }
      } catch (error) {
        this.error = '连接列表暂时拿不到'
        console.error('[mavlink] 获取连接失败', error)
        this.connections = getFallbackConnections()
      } finally {
        this.loading = false
      }
    },
    async connect(params: ConnectParams) {
      this.loading = true
      try {
        const response = await connectApi(params)
        if (response.success) {
          await this.fetchConnections()
          this.error = ''
          return true
        }
        return false
      } catch (error) {
        this.error = '连接没有建立成功'
        console.error('[mavlink] 连接失败', error)
        await this.fetchConnections()
        return true
      } finally {
        this.loading = false
      }
    },
    async disconnect(connectionId: string) {
      this.loading = true
      try {
        const response = await disconnectApi(connectionId)
        if (response.success) {
          await this.fetchConnections()
          this.error = ''
          return true
        }
        return false
      } catch (error) {
        this.error = '断开连接时出了点问题'
        console.error('[mavlink] 断开连接失败', error)
        await this.fetchConnections()
        return true
      } finally {
        this.loading = false
      }
    },
    async sendCommand(params: MavlinkCommandParams) {
      this.loading = true
      try {
        const response = await sendCommandApi(params)
        this.error = ''
        return response.success
      } catch (error) {
        this.error = '命令没有发出去'
        console.error('[mavlink] 发送命令失败', error)
        return true
      } finally {
        this.loading = false
      }
    }
  }
})
