<template>
  <div class="board-manager-container page-transition">
    <h1 class="gradient-title">📱 板子管理</h1>

    <div class="board-card">
      <div class="card-header">
        <h2>板子列表</h2>
        <div class="card-actions">
          <div class="connection-status">
            已连接：<span class="status-number">{{ connectedCount }}</span> / {{ boards.length }}
          </div>
          <el-button size="small" @click="createDialogVisible = true" type="primary">
            <el-icon><Plus /></el-icon>
            创建板子
          </el-button>
          <el-button size="small" @click="sendMessageDialogVisible = true" type="success">
            <el-icon><ChatLineRound /></el-icon>
            发送消息
          </el-button>
          <el-button size="small" @click="loadBoards">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>

      <div class="table-container">
        <el-table :data="boards" style="width: 100%" row-key="board_id" v-loading="loading">
          <el-table-column prop="board_id" label="板子ID" width="120" />
          <el-table-column prop="board_name" label="名称" min-width="150" />
          <el-table-column prop="address" label="IP" width="120" />
          <el-table-column prop="port" label="端口" width="80" />
          <el-table-column prop="connection" label="连接类型" width="100" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_connected ? 'success' : 'info'">
                {{ row.is_connected ? '已连接' : '未连接' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="sendCommand(row.board_id, 'TakePhoto')">拍照</el-button>
              <el-button size="small" @click="sendCommand(row.board_id, 'TakeOff')">起飞</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row.board_id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- 创建板子对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="创建板子"
      width="500px"
    >
      <el-form :model="createForm" :rules="createRules">
        <el-form-item label="板子ID" prop="board_id">
          <el-input v-model="createForm.board_id" placeholder="请输入板子ID" />
        </el-form-item>
        <el-form-item label="板子名称" prop="board_name">
          <el-input v-model="createForm.board_name" placeholder="请输入板子名称" />
        </el-form-item>
        <el-form-item label="板子类型" prop="board_type">
          <el-select v-model="createForm.board_type" placeholder="请选择板子类型">
            <el-option label="Drone" value="Drone" />
            <el-option label="Sensor" value="Sensor" />
          </el-select>
        </el-form-item>
        <el-form-item label="连接类型" prop="connection">
          <el-select v-model="createForm.connection" placeholder="请选择连接类型">
            <el-option label="TCP" value="TCP" />
            <el-option label="UDP" value="UDP" />
          </el-select>
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="createForm.address" placeholder="请输入地址" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input v-model="createForm.port" placeholder="请输入端口" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleCreate">创建</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 发送消息对话框 -->
    <el-dialog
      v-model="sendMessageDialogVisible"
      title="发送消息"
      width="500px"
    >
      <el-form :model="sendMessageForm" :rules="sendMessageRules">
        <el-form-item label="目标板子" prop="to_id">
          <el-select v-model="sendMessageForm.to_id" placeholder="请选择板子">
            <el-option v-for="board in boards" :key="board.board_id" :label="board.board_name || board.board_id" :value="board.board_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="命令" prop="command">
          <el-input v-model="sendMessageForm.command" placeholder="请输入命令" />
        </el-form-item>
        <el-form-item label="属性" prop="attribute">
          <el-input v-model="sendMessageForm.attribute" placeholder="请输入属性" />
        </el-form-item>
        <el-form-item label="数据">
          <el-input v-model="sendMessageForm.dataJson" placeholder="请输入数据（JSON格式）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="sendMessageDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSendMessage">发送</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, ChatLineRound } from '@element-plus/icons-vue'
import { getBoardListApi, sendMessageApi, createBoardApi, deleteBoardApi } from '@/api/board'

interface BoardRow {
  board_id: string
  board_name: string
  board_type: string
  connection: string
  address: string
  port: string
  is_connected: boolean
}

const boards = ref<BoardRow[]>([])
const loading = ref(false)

// 对话框状态
const createDialogVisible = ref(false)
const sendMessageDialogVisible = ref(false)

// 表单数据
const createForm = ref({
  board_id: '',
  board_name: '',
  board_type: 'Drone',
  connection: 'UDP',
  address: '0.0.0.0',
  port: '14550'
})

const sendMessageForm = ref({
  to_id: '',
  command: 'TakePhoto',
  attribute: 'Command',
  dataJson: '{}'
})

// 表单规则
const createRules = {
  board_id: [{ required: true, message: '请输入板子ID', trigger: 'blur' }],
  board_name: [{ required: true, message: '请输入板子名称', trigger: 'blur' }],
  board_type: [{ required: true, message: '请选择板子类型', trigger: 'change' }],
  connection: [{ required: true, message: '请选择连接类型', trigger: 'change' }],
  address: [{ required: true, message: '请输入地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }]
}

const sendMessageRules = {
  to_id: [{ required: true, message: '请选择板子', trigger: 'change' }],
  command: [{ required: true, message: '请输入命令', trigger: 'blur' }],
  attribute: [{ required: true, message: '请输入属性', trigger: 'blur' }]
}

const connectedCount = computed(() => boards.value.filter(b => b.is_connected).length)

// 加载板子列表
const loadBoards = async () => {
  loading.value = true
  try {
    const res = await getBoardListApi()
    if (res.success && res.data?.boards) {
      boards.value = res.data.boards.map((b) => ({
        board_id: b.boardId,
        board_name: b.boardName ?? '',
        board_type: String(b.boardType ?? ''),
        connection: String(b.boardStatus ?? ''),
        address: b.boardIp ?? '',
        port: b.boardPort ?? '',
        is_connected: b.isConnected
      }))
    }
  } catch (error) {
    ElMessage.error('加载板子列表失败')
  } finally {
    loading.value = false
  }
}

// 发送命令
const sendCommand = async (boardId: string, command: string) => {
  try {
    await sendMessageApi({
      to_id: boardId,
      to_type: 'Drone',
      command,
      attribute: 'Command',
      data: {}
    })
    ElMessage.success(`指令 ${command} 已发送`)
  } catch (error) {
    ElMessage.error('指令发送失败')
  }
}

// 创建板子
const handleCreate = async () => {
  try {
    await createBoardApi({
      board_id: createForm.value.board_id,
      board_name: createForm.value.board_name,
      board_type: createForm.value.board_type as 'Drone' | 'Sensor',
      connection: createForm.value.connection as 'TCP' | 'UDP',
      address: createForm.value.address,
      port: createForm.value.port
    })
    ElMessage.success('板子创建成功')
    createDialogVisible.value = false
    loadBoards()
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

// 删除板子
const handleDelete = async (boardId: string) => {
  await ElMessageBox.confirm('确定删除该板子吗？', '提示')
  try {
    await deleteBoardApi(boardId)
    ElMessage.success('删除成功')
    loadBoards()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

// 发送消息
const handleSendMessage = async () => {
  try {
    let data: Record<string, unknown> = {}
    try {
      data = JSON.parse(sendMessageForm.value.dataJson || '{}') as Record<string, unknown>
    } catch {
      ElMessage.error('数据必须是合法 JSON')
      return
    }
    await sendMessageApi({
      to_id: sendMessageForm.value.to_id,
      to_type: 'Drone',
      command: sendMessageForm.value.command,
      attribute: sendMessageForm.value.attribute,
      data
    })
    ElMessage.success('消息发送成功')
    sendMessageDialogVisible.value = false
  } catch (error) {
    ElMessage.error('消息发送失败')
  }
}

onMounted(() => {
  loadBoards()
})
</script>

<style scoped>
.board-manager-container {
  padding: 24px;
  min-height: 100vh;
  background: var(--bg-body);
  width: 100%;
  box-sizing: border-box;
  position: relative;
  z-index: 1;
}

.board-card {
  background: var(--bg-card);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  padding: 24px;
  position: relative;
  overflow: hidden;
  border: 1px solid var(--border-color);
  transition: all 0.3s ease;
}

/* 卡片左上角发光点 */
.board-card::before {
  content: '';
  position: absolute;
  top: 8px;
  left: 8px;
  width: 8px;
  height: 8px;
  background: linear-gradient(135deg, #2a5298, #6c9bd1);
  border-radius: 50%;
  box-shadow: 0 0 12px #2a5298;
  animation: pulse 2s ease-in-out infinite;
  z-index: 1;
}

@keyframes pulse {
  0%, 100% { box-shadow: 0 0 8px #2a5298; }
  50% { box-shadow: 0 0 20px #6c9bd1; }
}

/* 卡片悬浮效果 */
.board-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-hover);
  border-color: transparent;
}

/* 悬浮时渐变发光边框 */
.board-card:hover::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 1px;
  background: linear-gradient(135deg, #2a5298, #1e3c72);
  border-radius: 8px;
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
  animation: borderGlow 2s ease-in-out infinite alternate;
  z-index: 2;
}

@keyframes borderGlow {
  0% { opacity: 0.6; }
  100% { opacity: 1; }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
  position: relative;
  z-index: 3;
}

.card-header h2 {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.connection-status {
  color: var(--text-secondary);
  font-size: 14px;
}

.status-number {
  color: var(--primary);
  font-weight: 600;
  margin: 0 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

/* 表格样式 */
.table-container {
  max-height: 600px;
  overflow-y: auto;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  position: relative;
  z-index: 3;
}

:deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
  width: 100%;
}

:deep(.el-table__header-wrapper) {
  border-radius: 8px 8px 0 0;
  position: sticky;
  top: 0;
  z-index: 1;
}

:deep(.el-table th) {
  font-weight: 600;
  background-color: var(--bg-table-header);
  color: var(--text-primary);
}

:deep(.el-table td) {
  font-size: 14px;
  color: var(--text-secondary);
}

/* 滚动条样式 */
.table-container::-webkit-scrollbar {
  width: 8px;
}

.table-container::-webkit-scrollbar-track {
  background: #0d1321;
  border-radius: 4px;
}

.table-container::-webkit-scrollbar-thumb {
  background: #2a5298;
  border-radius: 4px;
}

.table-container::-webkit-scrollbar-thumb:hover {
  background: #1e3c72;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .board-manager-container {
    padding: 16px;
  }
  
  .board-card {
    padding: 16px;
  }
  
  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .card-actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>

