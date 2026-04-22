import { defineStore } from 'pinia'
import { getBoardListApi, createBoardApi, sendMessageApi, stopBoardApi } from '@/api/board'
import type { BoardInfo, CreateBoardParams, SendMessageParams } from '@/types/board'

export type BoardConnectionType = 'TCP' | 'UDP' | (string & {})

interface BoardState {
  boards: BoardInfo[]
  loading: boolean
}

export const useBoardStore = defineStore('board', {
  state: (): BoardState => ({
    boards: [],
    loading: false
  }),
  actions: {
    async fetchBoards() {
      this.loading = true
      try {
        const res = await getBoardListApi()
        if (res.success && res.data?.boards) {
          this.boards = res.data.boards
        }
      } catch (error) {
        console.log('获取板子列表失败，使用模拟数据')
        this.boards = [
          {
            boardId: 'drone_001',
            boardName: 'Drone Board',
            boardIp: '0.0.0.0',
            boardPort: '14550',
            boardStatus: 'TCP',
            isConnected: true
          }
        ]
      } finally {
        this.loading = false
      }
    },
    async createBoard(params: CreateBoardParams) {
      try {
        const res = await createBoardApi(params)
        if (res.success) {
          await this.fetchBoards()
          return true
        }
        return false
      } catch (error) {
        console.log('创建板子失败')
        return false
      }
    },
    async sendMessage(params: SendMessageParams) {
      try {
        const res = await sendMessageApi(params)
        return res.success
      } catch (error) {
        console.log('发送消息失败')
        return false
      }
    },
    async stopBoard() {
      try {
        const res = await stopBoardApi()
        if (res.success) {
          await this.fetchBoards()
          return true
        }
        return false
      } catch (error) {
        console.log('停止板子失败')
        return false
      }
    }
  }
})

