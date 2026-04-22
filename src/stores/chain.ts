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

function makeId() {
  return `${Date.now().toString(36)}-${Math.random().toString(16).slice(2)}`
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
        const res = await getChainListApi()
        if (res.success && res.data?.chains) {
          this.chains = res.data.chains
        }
      } catch (error) {
        console.log('获取任务链列表失败，使用模拟数据')
        this.chains = []
      } finally {
        this.loading = false
      }
    },
    async createChain(name: string, description?: string) {
      try {
        const res = await createChainApi({ name, description })
        if (res.success) {
          await this.fetchChains()
          return true
        }
        return false
      } catch (error) {
        console.log('创建任务链失败')
        return false
      }
    },
    async fetchChainDetail(id: string) {
      try {
        const res = await getChainApi(id)
        if (res.success && res.data?.chain) {
          this.currentChain = res.data.chain
          this.nodes = res.data.chain.nodes
        }
      } catch (error) {
        console.log('获取任务链详情失败')
      }
    },
    setNodes(nodes: ChainNode[]) {
      this.nodes = nodes
    },
    reorderNodesByIds(ids: string[]) {
      const map = new Map(this.nodes.map(n => [n.id, n]))
      const next: ChainNode[] = []

      for (const id of ids) {
        const node = map.get(id)
        if (node) next.push(node)
      }

      if (next.length !== this.nodes.length) {
        const picked = new Set(next.map(n => n.id))
        for (const n of this.nodes) {
          if (!picked.has(n.id)) next.push(n)
        }
      }

      this.nodes = next
    },
    async addNode(nodeType: ChainNodeType, params: Record<string, unknown>) {
      if (this.currentChain) {
        try {
          const res = await addNodeApi(this.currentChain.id, { nodeType, params })
          if (res.success) {
            await this.fetchChainDetail(this.currentChain.id)
            return true
          }
        } catch (error) {
          console.log('添加节点失败')
        }
      }

      this.nodes.push({
        id: makeId(),
        nodeType,
        params
      })
      return true
    },
    async removeNode(nodeId: string) {
      if (this.currentChain) {
        try {
          const res = await deleteNodeApi(this.currentChain.id, nodeId)
          if (res.success) {
            await this.fetchChainDetail(this.currentChain.id)
            return true
          }
        } catch (error) {
          console.log('删除节点失败')
        }
      }

      this.nodes = this.nodes.filter(n => n.id !== nodeId)
      return true
    },
    async startChain(chainId: string) {
      try {
        const res = await startChainApi(chainId)
        if (res.success) {
          await this.fetchChains()
          return true
        }
        return false
      } catch (error) {
        console.log('启动任务链失败')
        return false
      }
    },
    async stopChain(chainId: string) {
      try {
        const res = await stopChainApi(chainId)
        if (res.success) {
          await this.fetchChains()
          return true
        }
        return false
      } catch (error) {
        console.log('停止任务链失败')
        return false
      }
    }
  }
})

