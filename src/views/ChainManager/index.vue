<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import Sortable from 'sortablejs'
import { storeToRefs } from 'pinia'
import { ElMessage, ElDialog, ElForm, ElFormItem, ElInput, ElSelect, ElOption, ElButton, ElTable, ElTableColumn, ElTag, ElIcon } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useChainStore } from '@/stores/chain'
import type { ChainNodeType } from '@/types/chain'

const router = useRouter()

const chainStore = useChainStore()
const { nodes, chains, currentChain, loading } = storeToRefs(chainStore)

// 对话框状态
const addNodeDialogVisible = ref(false)

// 表单数据
const addNodeForm = ref({
  nodeType: 'takeoff',
  params: {
    altitude: 10,
    lat: 0,
    lng: 0,
    alt: 0,
    seconds: 0
  }
})

// 表单规则
const addNodeRules = {
  nodeType: [{ required: true, message: '请选择节点类型', trigger: 'change' }]
}

// 计算属性
const currentChainName = computed(() => currentChain.value?.name || '当前任务链')

// 生命周期
onMounted(() => {
  chainStore.fetchChains()
})

// 拖拽相关
const listRef = ref<HTMLDivElement | null>(null)
let sortable: Sortable | null = null

function syncOrderFromDom() {
  const el = listRef.value
  if (!el) return
  const ids = Array.from(el.querySelectorAll<HTMLElement>('[data-node-id]')).map(n => {
    return n.getAttribute('data-node-id') || ''
  })
  chainStore.reorderNodesByIds(ids.filter(Boolean))
}

onMounted(() => {
  if (!listRef.value) return
  sortable = Sortable.create(listRef.value, {
    animation: 150,
    ghostClass: 'chain-ghost',
    onEnd: () => {
      syncOrderFromDom()
    }
  })
})

onBeforeUnmount(() => {
  sortable?.destroy()
  sortable = null
})

// 方法

async function handleSelectChain(chainId: string) {
  await chainStore.fetchChainDetail(chainId)
  ElMessage.success('切换任务链成功')
}

async function handleAddNode() {
  try {
    const success = await chainStore.addNode(addNodeForm.value.nodeType as ChainNodeType, addNodeForm.value.params)
    if (success) {
      ElMessage.success('添加节点成功')
      addNodeDialogVisible.value = false
      // 重置表单
      addNodeForm.value = {
        nodeType: 'takeoff',
        params: {
          altitude: 10,
          lat: 0,
          lng: 0,
          alt: 0,
          seconds: 0
        }
      }
    } else {
      ElMessage.error('添加节点失败')
    }
  } catch (error) {
    ElMessage.error('添加节点失败')
  }
}

async function handleRemoveNode(id: string) {
  try {
    const success = await chainStore.removeNode(id)
    if (success) {
      ElMessage.success('删除节点成功')
    } else {
      ElMessage.error('删除节点失败')
    }
  } catch (error) {
    ElMessage.error('删除节点失败')
  }
}

async function handleStartChain() {
  if (!currentChain.value) {
    ElMessage.warning('请先选择一个任务链')
    return
  }
  try {
    const success = await chainStore.startChain(currentChain.value.id)
    if (success) {
      ElMessage.success('启动任务链成功')
      await chainStore.fetchChains()
    } else {
      ElMessage.error('启动任务链失败')
    }
  } catch (error) {
    ElMessage.error('启动任务链失败')
  }
}

async function handleStopChain() {
  if (!currentChain.value) {
    ElMessage.warning('请先选择一个任务链')
    return
  }
  try {
    const success = await chainStore.stopChain(currentChain.value.id)
    if (success) {
      ElMessage.success('停止任务链成功')
      await chainStore.fetchChains()
    } else {
      ElMessage.error('停止任务链失败')
    }
  } catch (error) {
    ElMessage.error('停止任务链失败')
  }
}

function handleCreateChain() {
  router.push('/chain-create')
}
</script>

<template>
  <div class="page page-transition">
    <div class="page-header">
      <h2 class="title">任务链管理</h2>
      <div class="actions">
        <el-button size="small" @click="handleCreateChain" type="primary"><el-icon><Plus /></el-icon> 新建任务链</el-button>
        <el-button size="small" @click="addNodeDialogVisible = true" type="success">添加节点</el-button>
        <el-button size="small" @click="handleStartChain" type="warning">启动任务链</el-button>
        <el-button size="small" @click="handleStopChain" type="danger">停止任务链</el-button>
      </div>
    </div>

    <div class="chain-section">
      <h3 class="section-title">任务链列表</h3>
      <el-table :data="chains" style="width: 100%" row-key="id" v-loading="loading">
        <el-table-column prop="name" label="任务链名称" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag :type="row.status === 'running' ? 'success' : row.status === 'completed' ? 'info' : row.status === 'failed' ? 'danger' : 'warning'">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作">
          <template #default="{ row }">
            <el-button size="small" @click="handleSelectChain(row.id)">选择</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="node-section">
      <h3 class="section-title">{{ currentChainName }} - 节点列表</h3>
      <div ref="listRef" class="node-list" aria-label="chain-nodes">
        <div
          v-for="n in nodes"
          :key="n.id"
          class="node-item"
          :data-node-id="n.id"
        >
          <div class="node-top">
            <el-tag type="info">{{ n.nodeType }}</el-tag>
            <span class="node-id">{{ n.id }}</span>
          </div>
          <div class="node-bottom">
            <div class="node-params">
              <div class="muted">params</div>
              <pre class="pre">{{ JSON.stringify(n.params, null, 2) }}</pre>
            </div>
            <div class="node-buttons">
              <el-button size="small" type="danger" link @click="handleRemoveNode(n.id)">删除</el-button>
            </div>
          </div>
        </div>
        <div v-if="nodes.length === 0" class="empty">
          暂无节点，请添加节点
        </div>
      </div>
    </div>



    <!-- 添加节点对话框 -->
    <el-dialog
      v-model="addNodeDialogVisible"
      title="添加节点"
      width="500px"
    >
      <el-form :model="addNodeForm" :rules="addNodeRules" ref="addNodeFormRef">
        <el-form-item label="节点类型" prop="nodeType">
          <el-select v-model="addNodeForm.nodeType" placeholder="请选择节点类型">
            <el-option label="起飞" value="takeoff" />
            <el-option label="降落" value="land" />
            <el-option label="移动" value="move" />
            <el-option label="返航" value="return" />
            <el-option label="等待" value="wait" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="参数">
          <el-input v-if="addNodeForm.nodeType === 'takeoff'" v-model.number="addNodeForm.params.altitude" type="number" placeholder="请输入起飞高度" />
          <div v-else-if="addNodeForm.nodeType === 'move'" class="params-grid">
            <el-input v-model.number="addNodeForm.params.lat" type="number" placeholder="纬度" />
            <el-input v-model.number="addNodeForm.params.lng" type="number" placeholder="经度" />
            <el-input v-model.number="addNodeForm.params.alt" type="number" placeholder="高度" />
          </div>
          <el-input v-else-if="addNodeForm.nodeType === 'wait'" v-model.number="addNodeForm.params.seconds" type="number" placeholder="请输入等待时间（秒）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="addNodeDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleAddNode">添加</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page {
  color: var(--text-primary);
  padding: 24px;
  min-height: 100vh;
  background: var(--bg-body);
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  flex-wrap: wrap;
  gap: 10px;
}

.title {
  font-size: 18px;
  margin: 0;
  font-weight: 700;
  color: var(--text-primary);
}

.actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.chain-section {
  margin-bottom: 24px;
}

.chain-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.chain-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 20px;
  transition: all 0.2s;
}

.chain-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-hover);
}

.chain-name {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.node-section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 16px;
  margin: 0 0 12px 0;
  font-weight: 600;
  color: var(--text-primary);
}

.node-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.node-item {
  border: 1px solid var(--border-color);
  background: var(--bg-card);
  border-radius: var(--radius-sm);
  padding: 14px;
  cursor: grab;
  transition: all 0.3s;
}

.node-item:hover {
  border-color: var(--primary);
  box-shadow: 0 4px 12px rgba(30, 136, 229, 0.1);
}

.node-item:active {
  cursor: grabbing;
}

.node-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 10px;
}

.node-id {
  font-size: 12px;
  color: var(--text-secondary);
}

.node-bottom {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
}

.node-params {
  flex: 1;
}

.muted {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.pre {
  margin: 0;
  font-size: 12px;
  line-height: 1.4;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 120px;
  overflow-y: auto;
  padding: 8px;
  background: var(--bg-body);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
}

.node-buttons {
  width: 90px;
  display: flex;
  justify-content: flex-end;
}

.chain-ghost {
  opacity: 0.45;
  background: rgba(30, 136, 229, 0.2);
  border: 1px dashed rgba(30, 136, 229, 0.5);
}

.empty {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
  border: 1px dashed var(--border-color);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
}

.params-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 10px;
}

/* 表格样式 */
:deep(.el-table) {
  --el-table-bg-color: var(--bg-card);
  --el-table-border-color: var(--border-color);
  --el-table-header-bg-color: var(--bg-table-header);
  --el-table-header-text-color: var(--text-primary);
  --el-table-row-hover-bg-color: #F5F7FA;
  --el-table-text-color: var(--text-secondary);
}

/* 对话框样式 */
:deep(.el-dialog) {
  --el-dialog-bg-color: var(--bg-card);
  --el-dialog-border-color: var(--border-color);
  --el-dialog-header-text-color: var(--text-primary);
}

:deep(.el-form-item__label) {
  color: var(--text-primary);
}

:deep(.el-input__wrapper) {
  background: var(--bg-body) !important;
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

:deep(.el-input input) {
  color: var(--text-primary);
  caret-color: var(--primary);
}

:deep(.el-select__input) {
  color: var(--text-primary);
}

:deep(.el-select__placeholder) {
  color: var(--text-secondary);
}

:deep(.el-option) {
  color: var(--text-primary);
}

:deep(.el-select-dropdown) {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
}

:deep(.el-option:hover) {
  background: #F5F7FA;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>

