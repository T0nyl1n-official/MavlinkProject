<template>
  <div class="chain-create-container page-transition">
    <div class="chain-create-card">
      <h1 class="chain-create-title gradient-title">创建任务链</h1>
      
      <!-- 任务链名称 -->
      <el-form-item label="任务链名称" class="chain-name-form">
        <el-input
          v-model="chainName"
          placeholder="请输入任务链名称"
          clearable
          class="chain-name-input"
        />
      </el-form-item>
      
      <!-- 任务节点列表 -->
      <div class="nodes-section">
        <h2 class="nodes-title">📋 任务节点列表</h2>
        
        <div class="nodes-list" ref="nodesListRef" v-if="nodes.length > 0">
          <div 
            v-for="(node, index) in nodes" 
            :key="node.id"
            class="node-item"
          >
            <div class="node-header">
              <span class="node-index">{{ index + 1 }}.</span>
              <span class="node-type">{{ getNodeTypeName(node.type) }}</span>
              <div class="node-actions">
                <el-button 
                  type="danger" 
                  size="small" 
                  circle
                  @click="removeNode(index)"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
                <el-button 
                  type="primary" 
                  size="small" 
                  circle
                  @click="moveNodeUp(index)"
                  :disabled="index === 0"
                >
                  <el-icon><ArrowUp /></el-icon>
                </el-button>
                <el-button 
                  type="primary" 
                  size="small" 
                  circle
                  @click="moveNodeDown(index)"
                  :disabled="index === nodes.length - 1"
                >
                  <el-icon><ArrowDown /></el-icon>
                </el-button>
              </div>
            </div>
            
            <!-- 节点参数 -->
            <div class="node-params">
              <template v-if="node.type === 'takeoff'">
                <el-form-item label="高度" size="small">
                  <el-input-number 
                    v-model="node.params.altitude" 
                    :min="1" 
                    :max="1000" 
                    :step="1"
                    placeholder="高度(米)"
                  />
                </el-form-item>
              </template>
              
              <template v-if="node.type === 'move'">
                <el-form-item label="纬度" size="small">
                  <el-input 
                    v-model="node.params.lat" 
                    type="number" 
                    placeholder="纬度"
                  />
                </el-form-item>
                <el-form-item label="经度" size="small">
                  <el-input 
                    v-model="node.params.lng" 
                    type="number" 
                    placeholder="经度"
                  />
                </el-form-item>
                <el-form-item label="高度" size="small">
                  <el-input-number 
                    v-model="node.params.alt" 
                    :min="1" 
                    :max="1000" 
                    :step="1"
                    placeholder="高度(米)"
                  />
                </el-form-item>
              </template>
              
              <template v-if="node.type === 'take_photo'">
                <el-form-item label="照片数量" size="small">
                  <el-input-number 
                    v-model="node.params.count" 
                    :min="1" 
                    :max="100" 
                    :step="1"
                    placeholder="照片数量"
                  />
                </el-form-item>
              </template>
              
              <template v-if="node.type === 'hover'">
                <el-form-item label="时长" size="small">
                  <el-input-number 
                    v-model="node.params.duration" 
                    :min="1" 
                    :max="3600" 
                    :step="1"
                    placeholder="时长(秒)"
                  />
                </el-form-item>
              </template>
            </div>
          </div>
        </div>
        
        <div v-else class="empty-nodes">
          <p>暂无任务节点，请添加</p>
        </div>
      </div>
      
      <!-- 添加节点按钮 -->
      <div class="add-node-section">
        <el-dropdown @command="addNode">
          <el-button type="primary" class="add-node-btn">
            <el-icon><Plus /></el-icon>
            添加任务节点
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="takeoff">起飞</el-dropdown-item>
              <el-dropdown-item command="move">移动到位置</el-dropdown-item>
              <el-dropdown-item command="take_photo">拍照</el-dropdown-item>
              <el-dropdown-item command="hover">悬停</el-dropdown-item>
              <el-dropdown-item command="land">降落</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      
      <!-- 操作按钮 -->
      <div class="action-buttons">
        <el-button 
          type="primary" 
          class="send-btn"
          @click="sendChain"
          :loading="loading"
          :disabled="!chainName || nodes.length === 0"
        >
          <el-icon v-if="!loading"><Check /></el-icon>
          发送任务
        </el-button>
        <el-button 
          type="info" 
          class="clear-btn"
          @click="clearAll"
          :disabled="nodes.length === 0"
        >
          <el-icon><DocumentRemove /></el-icon>
          清空
        </el-button>
      </div>
      
      <!-- 成功提示 -->
      <el-dialog
        v-model="successDialogVisible"
        title="任务链创建成功"
        width="400px"
      >
        <div class="success-content">
          <div class="success-icon">
            <el-icon class="success-icon-large"><Check /></el-icon>
          </div>
          <p class="success-message">任务链已发送给中控</p>
          <p class="chain-id">任务链ID: {{ successChainId }}</p>
        </div>
        <template #footer>
          <span class="dialog-footer">
            <el-button @click="successDialogVisible = false">关闭</el-button>
            <el-button type="primary" @click="createNewChain">再创建一个</el-button>
          </span>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete, ArrowUp, ArrowDown, DocumentRemove, Check } from '@element-plus/icons-vue'
import { createChainApi, addNodeApi, startChainApi } from '@/api/chain'
import Sortable from 'sortablejs'

// 任务链名称
const chainName = ref('')
// 加载状态
const loading = ref(false)
// 成功对话框
const successDialogVisible = ref(false)
// 成功创建的任务链ID
const successChainId = ref('')
// 节点列表DOM引用
const nodesListRef = ref<HTMLElement | null>(null)

// 任务节点接口
interface TaskNode {
  id: string
  type: string
  params: any
}

// 节点ID计数器
let nodeIdCounter = 0
// 任务节点列表
const nodes = ref<TaskNode[]>([])

// 节点类型默认参数
const nodeTypeDefaults = {
  takeoff: { altitude: 10 },
  move: { lat: 40.7128, lng: -74.0060, alt: 10 },
  take_photo: { count: 5 },
  hover: { duration: 10 },
  land: {}
}

// 生成节点ID
function generateNodeId(): string {
  return `node_${Date.now()}_${nodeIdCounter++}`
}

// Sortable实例
let sortableInstance: Sortable | null = null

// 页面加载时初始化拖拽
onMounted(() => {
  if (nodesListRef.value) {
    sortableInstance = Sortable.create(nodesListRef.value, {
      animation: 200,
      handle: '.node-item',
      ghostClass: 'sortable-ghost',
      onEnd: (evt) => {
        const { oldIndex, newIndex } = evt
        if (oldIndex !== undefined && newIndex !== undefined && oldIndex !== newIndex) {
          const movedNode = nodes.value.splice(oldIndex, 1)[0]
          nodes.value.splice(newIndex, 0, movedNode)
          ElMessage.success('节点顺序已更新')
        }
      }
    })
  }
})

// 页面卸载时清理
onUnmounted(() => {
  sortableInstance?.destroy()
  sortableInstance = null
})

// 获取节点类型的中文名称
function getNodeTypeName(type: string): string {
  const typeMap: Record<string, string> = {
    takeoff: '起飞',
    move: '移动到位置',
    take_photo: '拍照',
    hover: '悬停',
    land: '降落'
  }
  return typeMap[type] || type
}

// 添加新节点
function addNode(type: string) {
  const newNode: TaskNode = {
    id: generateNodeId(),
    type,
    params: JSON.parse(JSON.stringify(nodeTypeDefaults[type as keyof typeof nodeTypeDefaults]))
  }
  nodes.value.push(newNode)
}

// 删除节点
function removeNode(index: number) {
  const newNodes = nodes.value.filter((_, i) => i !== index)
  nodes.value = newNodes
  console.log('删除后节点数量:', nodes.value.length)
  ElMessage.success('节点已删除')
  
  // 重新初始化 Sortable 实例
  if (sortableInstance) {
    sortableInstance.destroy()
    sortableInstance = null
  }
  
  setTimeout(() => {
    if (nodesListRef.value) {
      sortableInstance = Sortable.create(nodesListRef.value, {
        animation: 200,
        handle: '.node-item',
        ghostClass: 'sortable-ghost',
        onEnd: (evt) => {
          const { oldIndex, newIndex } = evt
          if (oldIndex !== undefined && newIndex !== undefined && oldIndex !== newIndex) {
            const movedNode = nodes.value.splice(oldIndex, 1)[0]
            nodes.value.splice(newIndex, 0, movedNode)
            ElMessage.success('节点顺序已更新')
          }
        }
      })
    }
  }, 50)
}

// 上移节点
function moveNodeUp(index: number) {
  if (index > 0) {
    const temp = nodes.value[index]
    nodes.value[index] = nodes.value[index - 1]
    nodes.value[index - 1] = temp
    ElMessage.success('节点顺序已更新')
  }
}

// 下移节点
function moveNodeDown(index: number) {
  if (index < nodes.value.length - 1) {
    const temp = nodes.value[index]
    nodes.value[index] = nodes.value[index + 1]
    nodes.value[index + 1] = temp
    ElMessage.success('节点顺序已更新')
  }
}

// 清空所有内容
function clearAll() {
  nodes.value = []
  chainName.value = ''
  
  // 清理 Sortable 实例
  if (sortableInstance) {
    sortableInstance.destroy()
    sortableInstance = null
  }
}

// 发送任务链
const sendChain = async () => {
  if (!chainName.value) {
    ElMessage.warning('请输入任务链名称')
    return
  }
  if (nodes.value.length === 0) {
    ElMessage.warning('请至少添加一个任务节点')
    return
  }

  loading.value = true

  try {
    // 创建任务链
    const createRes = await createChainApi({ name: chainName.value })
    const chainId = createRes.data?.chain_id
    if (!chainId) {
      throw new Error('未返回任务链ID')
    }

    // 批量添加节点
    for (const node of nodes.value) {
      await addNodeApi(chainId, {
        nodeType: node.type,
        params: node.params
      })
    }

    // 启动任务链
    await startChainApi(chainId)

    // 显示成功信息
    ElMessage.success(`✅ 任务链已发送给中控！ID: ${chainId}`)
    successChainId.value = chainId
    successDialogVisible.value = true

  } catch (error: any) {
    ElMessage.error(`发送失败: ${error.message || '未知错误'}`)
  } finally {
    loading.value = false
  }
}

// 创建新任务链
function createNewChain() {
  clearAll()
  successDialogVisible.value = false
}
</script>

<style scoped>
.chain-create-container {
  min-height: 100vh;
  background: var(--bg-body);
  padding: 32px;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

.chain-create-card {
  background: var(--bg-card);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  padding: 48px;
  width: 100%;
  max-width: 800px;
  transition: all 0.3s ease;
}

.chain-create-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.chain-create-title {
  text-align: center;
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 40px;
  letter-spacing: 2px;
}

.chain-name-form {
  margin-bottom: 32px;
}

.chain-name-input {
  width: 100%;
  font-size: 1.1rem;
}

.nodes-section {
  margin-bottom: 32px;
}

.nodes-title {
  font-size: 1.3rem;
  font-weight: 500;
  margin-bottom: 24px;
  color: var(--text-primary);
}

.nodes-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.node-item {
  background: var(--bg-body);
  border-radius: 6px;
  padding: 20px;
  transition: all 0.3s ease;
}

.node-item:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.node-item.sortable-ghost {
  opacity: 0.5;
  background: var(--primary-soft);
  border: 2px dashed var(--primary);
}

.node-item.sortable-chosen {
  background: var(--bg-body);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.node-index {
  font-weight: 600;
  color: var(--primary);
  margin-right: 12px;
  font-size: 1.1rem;
}

.node-type {
  font-weight: 500;
  color: var(--text-primary);
  font-size: 1.1rem;
}

.node-actions {
  display: flex;
  gap: 8px;
}

.node-params {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--primary-soft);
}

.empty-nodes {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
  background: var(--bg-body);
  border-radius: 6px;
}

.add-node-section {
  margin-bottom: 32px;
}

.add-node-btn {
  width: 100%;
  height: 48px;
  font-size: 1.1rem;
  transition: all 0.3s ease;
}

.add-node-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.action-buttons {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.send-btn, .clear-btn {
  min-width: 160px;
  height: 48px;
  font-size: 1.1rem;
  transition: all 0.3s ease;
}

.send-btn:hover, .clear-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.success-content {
  text-align: center;
  padding: 24px 0;
}

.success-icon {
  margin-bottom: 16px;
}

.success-icon-large {
  font-size: 48px;
  color: var(--primary);
}

.success-message {
  font-size: 1.2rem;
  font-weight: 500;
  margin-bottom: 16px;
  color: var(--text-primary);
}

.chain-id {
  font-size: 1rem;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .chain-create-container {
    padding: 16px;
  }
  
  .chain-create-card {
    padding: 24px;
  }
  
  .action-buttons {
    flex-direction: column;
  }
  
  .send-btn, .clear-btn {
    width: 100%;
  }
}
</style>