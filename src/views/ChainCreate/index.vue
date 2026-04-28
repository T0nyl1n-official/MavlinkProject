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
              <div v-if="getNodeSchema(node.type).fields.length > 0" class="params-grid">
                <el-form-item
                  v-for="field in getNodeSchema(node.type).fields"
                  :key="`${node.id}-${field.key}`"
                  :label="field.label"
                  size="small"
                  class="param-item"
                >
                  <el-input-number
                    v-if="field.component === 'number'"
                    v-model="node.params[field.key]"
                    :min="field.min"
                    :max="field.max"
                    :step="field.step ?? 1"
                    :placeholder="field.placeholder"
                    class="param-control"
                  />
                  <el-select
                    v-else-if="field.component === 'select'"
                    v-model="node.params[field.key]"
                    :placeholder="field.placeholder"
                    class="param-control"
                  >
                    <el-option
                      v-for="option in field.options || []"
                      :key="option.value"
                      :label="option.label"
                      :value="option.value"
                    />
                  </el-select>
                  <el-input
                    v-else
                    v-model="node.params[field.key]"
                    :placeholder="field.placeholder"
                    class="param-control"
                  />
                </el-form-item>
              </div>
              <div v-else class="empty-params">该节点无需额外参数</div>
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
              <el-dropdown-item
                v-for="schema in nodeSchemas"
                :key="schema.type"
                :command="schema.type"
              >
                {{ schema.label }}
              </el-dropdown-item>
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
import { nextTick, ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete, ArrowUp, ArrowDown, DocumentRemove, Check } from '@element-plus/icons-vue'
import { createChainApi, addNodeApi, startChainApi } from '@/api/chain'
import Sortable from 'sortablejs'
import type { ChainNodeType } from '@/types/chain'

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
  type: ChainNodeType
  params: Record<string, any>
}

interface FieldOption {
  label: string
  value: string
}

interface NodeField {
  key: string
  label: string
  component: 'number' | 'input' | 'select'
  placeholder: string
  min?: number
  max?: number
  step?: number
  options?: FieldOption[]
}

interface NodeSchema {
  type: ChainNodeType
  label: string
  defaults: Record<string, any>
  fields: NodeField[]
}

let nodeIdCounter = 0
const nodes = ref<TaskNode[]>([])

const flightModeOptions: FieldOption[] = [
  { label: 'STABILIZE', value: 'STABILIZE' },
  { label: 'ALT_HOLD', value: 'ALT_HOLD' },
  { label: 'LOITER', value: 'LOITER' },
  { label: 'RTL', value: 'RTL' },
  { label: 'AUTO', value: 'AUTO' },
  { label: 'GUIDED', value: 'GUIDED' }
]

const nodeSchemas: NodeSchema[] = [
  {
    type: 'takeoff',
    label: '起飞',
    defaults: { altitude: 10, timeout: 30 },
    fields: [
      { key: 'altitude', label: '高度', component: 'number', placeholder: '起飞高度(米)', min: 1, max: 1000 },
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 3600 }
    ]
  },
  {
    type: 'land',
    label: '降落',
    defaults: { latitude: 22.543123, longitude: 114.052345, timeout: 60 },
    fields: [
      { key: 'latitude', label: '纬度', component: 'number', placeholder: '降落点纬度', min: -90, max: 90, step: 0.000001 },
      { key: 'longitude', label: '经度', component: 'number', placeholder: '降落点经度', min: -180, max: 180, step: 0.000001 },
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 3600 }
    ]
  },
  {
    type: 'goto',
    label: '飞往目标',
    defaults: { latitude: 22.543123, longitude: 114.052345, altitude: 20, timeout: 60 },
    fields: [
      { key: 'latitude', label: '纬度', component: 'number', placeholder: '目标纬度', min: -90, max: 90, step: 0.000001 },
      { key: 'longitude', label: '经度', component: 'number', placeholder: '目标经度', min: -180, max: 180, step: 0.000001 },
      { key: 'altitude', label: '高度', component: 'number', placeholder: '目标高度(米)', min: 1, max: 1000 },
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 3600 }
    ]
  },
  {
    type: 'goto_location',
    label: '飞往目标(兼容别名)',
    defaults: { latitude: 22.543123, longitude: 114.052345, altitude: 20, timeout: 60 },
    fields: [
      { key: 'latitude', label: '纬度', component: 'number', placeholder: '目标纬度', min: -90, max: 90, step: 0.000001 },
      { key: 'longitude', label: '经度', component: 'number', placeholder: '目标经度', min: -180, max: 180, step: 0.000001 },
      { key: 'altitude', label: '高度', component: 'number', placeholder: '目标高度(米)', min: 1, max: 1000 },
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 3600 }
    ]
  },
  {
    type: 'return_to_home',
    label: '返航',
    defaults: { timeout: 60 },
    fields: [
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 3600 }
    ]
  },
  {
    type: 'rtl',
    label: '返航(兼容别名)',
    defaults: { timeout: 60 },
    fields: [
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 3600 }
    ]
  },
  {
    type: 'survey',
    label: '区域侦察',
    defaults: { latitude: 22.543123, longitude: 114.052345, radius: 50, duration: 30, timeout: 120 },
    fields: [
      { key: 'latitude', label: '中心纬度', component: 'number', placeholder: '中心纬度', min: -90, max: 90, step: 0.000001 },
      { key: 'longitude', label: '中心经度', component: 'number', placeholder: '中心经度', min: -180, max: 180, step: 0.000001 },
      { key: 'radius', label: '半径', component: 'number', placeholder: '侦察半径(米)', min: 1, max: 10000 },
      { key: 'duration', label: '时长', component: 'number', placeholder: '侦察时长(秒)', min: 1, max: 3600 },
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 7200 }
    ]
  },
  {
    type: 'survey_grid',
    label: '网格搜索',
    defaults: { latitude: 22.543123, longitude: 114.052345, width: 100, height: 100, altitude: 20, timeout: 180 },
    fields: [
      { key: 'latitude', label: '起点纬度', component: 'number', placeholder: '起点纬度', min: -90, max: 90, step: 0.000001 },
      { key: 'longitude', label: '起点经度', component: 'number', placeholder: '起点经度', min: -180, max: 180, step: 0.000001 },
      { key: 'width', label: '宽度', component: 'number', placeholder: '搜索宽度(米)', min: 1, max: 10000 },
      { key: 'height', label: '高度范围', component: 'number', placeholder: '搜索高度/范围(米)', min: 1, max: 10000 },
      { key: 'altitude', label: '飞行高度', component: 'number', placeholder: '飞行高度(米)', min: 1, max: 1000 },
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 7200 }
    ]
  },
  {
    type: 'orbit',
    label: '盘旋巡逻',
    defaults: { latitude: 22.543123, longitude: 114.052345, radius: 30, duration: 30, timeout: 120 },
    fields: [
      { key: 'latitude', label: '中心纬度', component: 'number', placeholder: '中心纬度', min: -90, max: 90, step: 0.000001 },
      { key: 'longitude', label: '中心经度', component: 'number', placeholder: '中心经度', min: -180, max: 180, step: 0.000001 },
      { key: 'radius', label: '盘旋半径', component: 'number', placeholder: '盘旋半径(米)', min: 1, max: 10000 },
      { key: 'duration', label: '盘旋时长', component: 'number', placeholder: '盘旋时长(秒)', min: 1, max: 3600 },
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 7200 }
    ]
  },
  {
    type: 'take_photo',
    label: '拍照',
    defaults: { timeout: 30 },
    fields: [
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 3600 }
    ]
  },
  {
    type: 'start_video',
    label: '开始录像',
    defaults: { timeout: 30 },
    fields: [
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 3600 }
    ]
  },
  {
    type: 'stop_video',
    label: '停止录像',
    defaults: { timeout: 30 },
    fields: [
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 3600 }
    ]
  },
  {
    type: 'set_mode',
    label: '设置模式',
    defaults: { mode: 'GUIDED', timeout: 30 },
    fields: [
      { key: 'mode', label: '飞行模式', component: 'select', placeholder: '请选择模式', options: flightModeOptions },
      { key: 'timeout', label: '超时', component: 'number', placeholder: '超时(秒)', min: 1, max: 3600 }
    ]
  }
]

const nodeSchemaMap = Object.fromEntries(nodeSchemas.map(schema => [schema.type, schema])) as Record<string, NodeSchema>

function generateNodeId(): string {
  return `node_${Date.now()}_${nodeIdCounter++}`
}

let sortableInstance: Sortable | null = null

function cloneDefaults<T>(value: T): T {
  return JSON.parse(JSON.stringify(value))
}

function swapNodes(currentIndex: number, targetIndex: number) {
  const currentNode = nodes.value[currentIndex]
  nodes.value[currentIndex] = nodes.value[targetIndex]
  nodes.value[targetIndex] = currentNode
  ElMessage.success('节点顺序已更新')
}

function resetComposer() {
  nodes.value = []
  chainName.value = ''
  destroySortable()
}

function getActionErrorMessage(error: any) {
  if (error?.response?.status === 401) {
    return '登录状态失效了，请重新登录后再试'
  }

  return error?.message || '任务没有发送成功'
}

function initSortable() {
  if (!nodesListRef.value || sortableInstance) {
    return
  }

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

function destroySortable() {
  sortableInstance?.destroy()
  sortableInstance = null
}

onMounted(() => {
  initSortable()
})

onUnmounted(() => {
  destroySortable()
})

function getNodeSchema(type: string): NodeSchema {
  return nodeSchemaMap[type] || {
    type,
    label: type,
    defaults: {},
    fields: []
  }
}

function getNodeTypeName(type: string): string {
  return getNodeSchema(type).label
}

async function addNode(type: string) {
  const schema = getNodeSchema(type)
  const node: TaskNode = {
    id: generateNodeId(),
    type: type as ChainNodeType,
    params: cloneDefaults(schema.defaults)
  }
  nodes.value.push(node)
  await nextTick()
  initSortable()
}

function removeNode(index: number) {
  nodes.value = nodes.value.filter((_, currentIndex) => currentIndex !== index)
  ElMessage.success('节点已删除')

  if (nodes.value.length === 0) {
    destroySortable()
  }
}

function moveNodeUp(index: number) {
  if (index > 0) {
    swapNodes(index, index - 1)
  }
}

function moveNodeDown(index: number) {
  if (index < nodes.value.length - 1) {
    swapNodes(index, index + 1)
  }
}

function clearAll() {
  resetComposer()
}

function validateNodeParams(node: TaskNode): string | null {
  const schema = getNodeSchema(node.type)

  for (const field of schema.fields) {
    const value = node.params[field.key]

    if (value === '' || value === null || value === undefined) {
      return `${schema.label} 节点缺少参数：${field.label}`
    }
  }

  return null
}

const sendChain = async () => {
  if (!chainName.value) {
    ElMessage.warning('请输入任务链名称')
    return
  }
  if (nodes.value.length === 0) {
    ElMessage.warning('请至少添加一个任务节点')
    return
  }

  for (const node of nodes.value) {
    const errorMessage = validateNodeParams(node)
    if (errorMessage) {
      ElMessage.warning(errorMessage)
      return
    }
  }

  loading.value = true

  try {
    const createdChain = await createChainApi({ name: chainName.value })
    const chainId = createdChain.data?.chain_id
    if (!chainId) {
      throw new Error('未返回任务链ID')
    }

    for (const node of nodes.value) {
      await addNodeApi(chainId, {
        nodeType: node.type,
        params: node.params
      })
    }

    await startChainApi(chainId)

    ElMessage.success(`任务链已下发，编号 ${chainId}`)
    successChainId.value = chainId
    successDialogVisible.value = true

  } catch (error: any) {
    ElMessage.error(getActionErrorMessage(error))
  } finally {
    loading.value = false
  }
}

function createNewChain() {
  resetComposer()
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

.params-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px 16px;
}

.param-item {
  margin-bottom: 0;
}

.param-control {
  width: 100%;
}

.empty-params {
  color: var(--text-secondary);
  font-size: 14px;
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