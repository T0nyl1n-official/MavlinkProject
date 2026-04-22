import request from './request'
import { mockService, config } from '@/utils/mockService'
import type { MockBoard } from '@/utils/mockService'
import type {
  BoardInfo,
  BoardInfoResponse,
  BoardListResponse,
  BoardType,
  CreateBoardParams,
  CreateBoardResponse,
  SendMessageParams,
  SendMessageResponse,
  StopBoardResponse
} from '@/types/board'

function mockBoardToBoardInfo(b: MockBoard): BoardInfo {
  return {
    boardId: b.board_id,
    boardName: b.board_name,
    boardType: b.board_type as BoardType,
    boardIp: b.address,
    boardPort: b.port,
    boardStatus: b.connection,
    isConnected: b.is_connected
  }
}

export function createBoardApi(params: CreateBoardParams): Promise<CreateBoardResponse> {
  if (!config.USE_REAL_API) {
    return mockService.createBoard({
      board_name: params.board_name ?? '',
      board_type: params.board_type,
      connection: params.connection,
      address: params.address,
      port: params.port
    }).then(board => ({
      code: 0,
      success: true,
      data: {
        board_id: board.board_id,
        message: 'Board created successfully'
      },
      message: 'Board created successfully'
    }))
  }
  return request.post('/api/board/create', params)
}

export function sendMessageApi(params: SendMessageParams): Promise<SendMessageResponse> {
  if (!config.USE_REAL_API) {
    return mockService.sendBoardCommand(params.to_id, params.command).then(() => ({
      code: 0,
      success: true,
      data: { message: 'Message sent successfully' },
      message: 'Message sent successfully'
    }))
  }
  return request.post('/api/board/send', params)
}

export function getBoardListApi(): Promise<BoardListResponse> {
  if (!config.USE_REAL_API) {
    return mockService.getBoards().then(boards => ({
      code: 0,
      success: true,
      data: {
        boards: boards.map(mockBoardToBoardInfo)
      },
      message: 'Board list retrieved successfully'
    }))
  }
  return request.get('/api/board/list')
}

export function getBoardInfoApi(boardID: string): Promise<BoardInfoResponse> {
  if (!config.USE_REAL_API) {
    return mockService.getBoards().then(boards => {
      const board = boards.find(b => b.board_id === boardID)
      const fallback: MockBoard = {
        board_id: boardID,
        board_name: 'Unknown Board',
        board_type: 'Unknown',
        connection: 'Unknown',
        address: 'Unknown',
        port: 'Unknown',
        is_connected: false
      }
      return {
        code: 0,
        success: true,
        data: {
          board: mockBoardToBoardInfo(board ?? fallback)
        },
        message: 'Board info retrieved successfully'
      }
    })
  }
  return request.get(`/api/board/info/${boardID}`)
}

export function stopBoardApi(): Promise<StopBoardResponse> {
  if (!config.USE_REAL_API) {
    return Promise.resolve({
      code: 0,
      success: true,
      data: { message: 'Board stopped successfully' },
      message: 'Board stopped successfully'
    })
  }
  return request.post('/api/board/stop')
}

export function deleteBoardApi(boardId: string): Promise<{ success: boolean; message: string }> {
  if (!config.USE_REAL_API) {
    return mockService.deleteBoard(boardId).then(() => ({
      code: 0,
      success: true,
      message: 'Board deleted successfully'
    }))
  }
  return request.delete(`/api/board/delete/${boardId}`)
}

/** 与旧命名兼容：发送板子指令 */
export const sendBoardCommandApi = sendMessageApi