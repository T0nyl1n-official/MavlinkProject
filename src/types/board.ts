import type { ApiResponse, JsonObject } from './api'

export type BoardConnectionType = 'TCP' | 'UDP' | (string & {})
export type BoardType = 'Drone' | 'Sensor' | (string & {})

export interface CreateBoardParams {
  board_id: string
  board_name?: string
  board_type: BoardType
  connection: BoardConnectionType
  address: string
  port: string
}

export interface BoardInfo extends JsonObject {
  boardId: string
  boardName?: string
  boardType?: BoardType
  boardIp?: string
  boardPort?: string
  boardStatus?: string
  isConnected: boolean
}

export interface SendMessageParams {
  to_id: string
  to_type: string
  command: string
  attribute: string
  data: JsonObject
}

export type BoardListResponse = ApiResponse<{ boards: BoardInfo[] }>
export type BoardInfoResponse = ApiResponse<{ board: BoardInfo }>
export type CreateBoardResponse = ApiResponse<{ board_id: string; message?: string }>
export type SendMessageResponse = ApiResponse<{ message: string }>
export type StopBoardResponse = ApiResponse<{ message: string }>