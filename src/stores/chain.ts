import { defineStore } from 'pinia'
import {
  getChainListApi,
  getChainApi,
  createChainApi,
  addNodeApi,
  deleteNodeApi,
  startChainApi,
  stopChainApi
} from '@/api/chain'
import type { ChainNode, ChainNodeType, ChainInfo } from '@/types/chain'

interface ChainState {
  chains: ChainInfo[]
  currentChain: ChainInfo | null
  nodes: ChainNode[]
  loading: boolean
}

function createLocalNodeId() {
  return `${Date.now().toString(36)}-${Math.random().toString(16).slice(2)}`
}

function logChainStoreError(action: string, error: unknown) {
  console.error(`[chain] ${action}`, error)
}

export const useChainStore = defineStore('chain', {
  state: (): ChainState => ({
    chains: [],
    currentChain: null,
    nodes: [],
    loading: false
  }),
  actions: {
    async fetchChains() {
      this.loading = true
      try {
        const response = await getChainListApi()
        if (response.success && response.data?.chains) {
          this.chains = response.data.chains
        }
      } catch (error) {
        logChainStoreError('获取任务链列表失败', error)
        this.chains = []
      } finally {
        this.loading = false
      }
    },
    async createChain(name: string, description?: string) {
      try {
        const response = await createChainApi({ name, description })
        if (response.success) {
          await this.fetchChains()
          return true
        }
        return false
      } catch (error) {
        logChainStoreError('创建任务链失败', error)
        return false
      }
    },
    async fetchChainDetail(chainId: string) {
      try {
        const response = await getChainApi(chainId)
        if (response.success && response.data?.chain) {
          this.currentChain = response.data.chain
          this.nodes = response.data.chain.nodes
        }
      } catch (error) {
        logChainStoreError('获取任务链详情失败', error)
      }
    },
    setNodes(chainNodes: ChainNode[]) {
      this.nodes = chainNodes
    },
    reorderNodesByIds(ids: string[]) {
      const nodeMap = new Map(this.nodes.map(node => [node.id, node]))
      const reorderedNodes: ChainNode[] = []

      for (const id of ids) {
        const node = nodeMap.get(id)
        if (node) reorderedNodes.push(node)
      }

      if (reorderedNodes.length !== this.nodes.length) {
        const appendedIds = new Set(reorderedNodes.map(node => node.id))
        for (const node of this.nodes) {
          if (!appendedIds.has(node.id)) reorderedNodes.push(node)
        }
      }

      this.nodes = reorderedNodes
    },
    async addNode(nodeType: ChainNodeType, params: Record<string, unknown>) {
      if (this.currentChain) {
        try {
          const response = await addNodeApi(this.currentChain.id, { nodeType, params })
          if (response.success) {
            await this.fetchChainDetail(this.currentChain.id)
            return true
          }
        } catch (error) {
          logChainStoreError('添加节点失败', error)
        }
      }

      this.nodes.push({
        id: createLocalNodeId(),
        nodeType,
        params
      })
      return true
    },
    async removeNode(nodeId: string) {
      if (this.currentChain) {
        try {
          const response = await deleteNodeApi(this.currentChain.id, nodeId)
          if (response.success) {
            await this.fetchChainDetail(this.currentChain.id)
            return true
          }
        } catch (error) {
          logChainStoreError('删除节点失败', error)
        }
      }

      this.nodes = this.nodes.filter(n => n.id !== nodeId)
      return true
    },
    async startChain(chainId: string) {
      try {
        const response = await startChainApi(chainId)
        if (response.success) {
          await this.fetchChains()
          return true
        }
        return false
      } catch (error) {
        logChainStoreError('启动任务链失败', error)
        return false
      }
    },
    async stopChain(chainId: string) {
      try {
        const response = await stopChainApi(chainId)
        if (response.success) {
          await this.fetchChains()
          return true
        }
        return false
      } catch (error) {
        logChainStoreError('停止任务链失败', error)
        return false
      }
    }
  }
})

