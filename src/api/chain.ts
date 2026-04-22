import request from './request'
import { mockService, config } from '@/utils/mockService'
import type { MockChain, MockNode } from '@/utils/mockService'
import type {
  AddNodeParams,
  AddNodeResponse,
  ChainDetailResponse,
  ChainInfo,
  ChainListResponse,
  ChainNode,
  ChainNodeType,
  ChainStatus,
  CreateChainParams,
  CreateChainResponse,
  DeleteNodeResponse,
  StartChainResponse,
  StopChainResponse
} from '@/types/chain'

function mockNodeToChainNode(n: MockNode): ChainNode {
  return {
    id: n.node_id,
    nodeType: n.node_type as ChainNodeType,
    params: n.parameters,
    status: 'pending'
  }
}

function mockChainToChainInfo(c: MockChain): ChainInfo {
  return {
    id: c.chain_id,
    name: c.chain_name,
    description: c.description,
    nodes: c.nodes.map(mockNodeToChainNode),
    status: c.status as ChainStatus,
    createdAt: c.created_at,
    updatedAt: c.created_at
  }
}

export function createChainApi(params: CreateChainParams): Promise<CreateChainResponse> {
  if (!config.USE_REAL_API) {
    return mockService.createChain(params.name, params.description || '').then(chain => ({
      code: 0,
      success: true,
      data: {
        chain_id: chain.chain_id
      },
      message: 'Chain created successfully'
    }))
  }
  return request.post('/api/chain/create', params)
}

export function getChainApi(id: string): Promise<ChainDetailResponse> {
  if (!config.USE_REAL_API) {
    return mockService.getChains().then(chains => {
      const chain = chains.find(c => c.chain_id === id)
      const fallback: MockChain = {
        chain_id: id,
        chain_name: 'Unknown Chain',
        description: '',
        nodes: [],
        status: 'idle',
        created_at: new Date().toISOString()
      }
      return {
        code: 0,
        success: true,
        data: {
          chain: mockChainToChainInfo(chain ?? fallback)
        },
        message: 'Chain retrieved successfully'
      }
    })
  }
  return request.get(`/api/chain/${id}`)
}

export function getChainListApi(): Promise<ChainListResponse> {
  if (!config.USE_REAL_API) {
    return mockService.getChains().then(chains => ({
      code: 0,
      success: true,
      data: {
        chains: chains.map(mockChainToChainInfo)
      },
      message: 'Chain list retrieved successfully'
    }))
  }
  return request.get('/api/chain/list')
}

export function addNodeApi(chainId: string, params: AddNodeParams): Promise<AddNodeResponse> {
  if (!config.USE_REAL_API) {
    return mockService.addNode(chainId, params.nodeType, 'Node', params.params).then(node => ({
      code: 0,
      success: true,
      data: {
        node_id: node.node_id
      },
      message: 'Node added successfully'
    }))
  }
  return request.post(`/api/chain/${chainId}/node/add`, params)
}

export function deleteNodeApi(chainId: string, nodeId: string): Promise<DeleteNodeResponse> {
  if (!config.USE_REAL_API) {
    return mockService.deleteNode(chainId, nodeId).then(() => ({
      code: 0,
      success: true,
      data: { message: 'Node deleted successfully' },
      message: 'Node deleted successfully'
    }))
  }
  return request.delete(`/api/chain/${chainId}/node/delete/${nodeId}`)
}

export function startChainApi(chainId: string): Promise<StartChainResponse> {
  if (!config.USE_REAL_API) {
    return mockService.startChain(chainId).then(() => ({
      code: 0,
      success: true,
      data: { status: 'running' },
      message: 'Chain started successfully'
    }))
  }
  return request.post(`/api/chain/${chainId}/start`)
}

export function stopChainApi(chainId: string): Promise<StopChainResponse> {
  if (!config.USE_REAL_API) {
    return mockService.stopChain(chainId).then(() => ({
      code: 0,
      success: true,
      data: { status: 'stopped' },
      message: 'Chain stopped successfully'
    }))
  }
  return request.post(`/api/chain/${chainId}/stop`)
}