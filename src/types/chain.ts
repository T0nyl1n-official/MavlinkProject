import type { ApiResponse, JsonObject } from './api'

export type ChainNodeType =
  | 'takeoff'
  | 'land'
  | 'goto'
  | 'goto_location'
  | 'return_to_home'
  | 'rtl'
  | 'survey'
  | 'survey_grid'
  | 'orbit'
  | 'take_photo'
  | 'start_video'
  | 'stop_video'
  | 'set_mode'
  | 'move'
  | 'return'
  | 'wait'
  | 'custom'
  | (string & {})
export type ChainStatus = 'idle' | 'running' | 'paused' | 'completed' | 'failed' | (string & {})

export interface ChainNode extends JsonObject {
  id: string
  nodeType: ChainNodeType
  params: JsonObject
  status?: 'pending' | 'running' | 'completed' | 'failed'
  result?: JsonObject
}

export interface ChainInfo extends JsonObject {
  id: string
  name: string
  description?: string
  nodes: ChainNode[]
  status: ChainStatus
  createdAt: string
  updatedAt: string
}

export interface CreateChainParams {
  name: string
  description?: string
  nodes?: ChainNode[]
}

export interface AddNodeParams {
  nodeType: ChainNodeType
  params: JsonObject
}

export type ChainListResponse = ApiResponse<{ chains: ChainInfo[] }>
export type ChainDetailResponse = ApiResponse<{ chain: ChainInfo }>
export type CreateChainResponse = ApiResponse<{ chain_id: string }>
export type AddNodeResponse = ApiResponse<{ node_id: string }>
export type DeleteNodeResponse = ApiResponse<{ message: string }>
export type StartChainResponse = ApiResponse<{ status: string }>
export type StopChainResponse = ApiResponse<{ status: string }>