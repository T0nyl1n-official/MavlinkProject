import { defineStore } from 'pinia'
import { getBoardListApi, createBoardApi, sendMessageApi, stopBoardApi } from '@/api/board'
import type { BoardInfo, CreateBoardParams, SendMessageParams } from '@/types/board'

export type BoardConnectionType = 'TCP' | 'UDP' | (string & {})

interface BoardState {
  boards: BoardInfo[]
  loading: boolean
}

function logBoardStoreError(action: string, error: unknown) {
  console.error(`[board] ${action}`, error)
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
        const response = await getBoardListApi()
        if (response.success && response.data?.boards) {
          this.boards = response.data.boards
        }
      } catch (error) {
        logBoardStoreError('获取板子列表失败', error)
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
        const response = await createBoardApi(params)
        if (response.success) {
          await this.fetchBoards()
          return true
        }
        return false
      } catch (error) {
        logBoardStoreError('创建板子失败', error)
        return false
      }
    },
    async sendMessage(params: SendMessageParams) {
      try {
        const response = await sendMessageApi(params)
        return response.success
      } catch (error) {
        logBoardStoreError('发送消息失败', error)
        return false
      }
    },
    async stopBoard() {
      try {
        const response = await stopBoardApi()
        if (response.success) {
          await this.fetchBoards()
          return true
        }
        return false
      } catch (error) {
        logBoardStoreError('停止板子失败', error)
        return false
      }
    }
  }
})

