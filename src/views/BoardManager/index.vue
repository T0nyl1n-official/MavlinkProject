<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, ElDialog, ElForm, ElFormItem, ElInput, ElSelect, ElOption, ElButton, ElTable, ElTableColumn, ElTag } from 'element-plus'
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

<template>
  <div class="page">
    <div class="page-header">
      <h2 class="title">板子管理</h2>
      <div class="actions">
        <div class="stats">
          已连接：<span class="num">{{ connectedCount }}</span> / {{ boards.length }}
        </div>
        <el-button size="small" @click="createDialogVisible = true" type="primary">创建板子</el-button>
        <el-button size="small" @click="sendMessageDialogVisible = true" type="success">发送消息</el-button>
        <el-button size="small" @click="loadBoards">刷新</el-button>
      </div>
    </div>

    <el-table :data="boards" style="width: 100%" row-key="board_id" v-loading="loading">
      <el-table-column prop="board_id" label="板子ID" />
      <el-table-column prop="board_name" label="名称" />
      <el-table-column prop="address" label="IP" />
      <el-table-column prop="port" label="端口" />
      <el-table-column prop="connection" label="连接类型" />
      <el-table-column label="状态">
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

<style scoped>
.page {
  color: rgba(255, 255, 255, 0.92);
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
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.stats {
  color: rgba(255, 255, 255, 0.75);
  display: flex;
  align-items: center;
  gap: 12px;
}

.num {
  color: #66aaff;
  font-weight: 700;
}

/* 表格样式 */
:deep(.el-table) {
  --el-table-bg-color: rgba(255, 255, 255, 0.03);
  --el-table-border-color: rgba(255, 255, 255, 0.08);
  --el-table-header-bg-color: rgba(255, 255, 255, 0.05);
  --el-table-header-text-color: rgba(255, 255, 255, 0.85);
  --el-table-row-hover-bg-color: rgba(255, 255, 255, 0.05);
  --el-table-text-color: rgba(255, 255, 255, 0.9);
}

/* 对话框样式 */
:deep(.el-dialog) {
  --el-dialog-bg-color: rgba(21, 21, 30, 0.95);
  --el-dialog-border-color: rgba(255, 255, 255, 0.12);
  --el-dialog-header-text-color: rgba(255, 255, 255, 0.9);
}

:deep(.el-form-item__label) {
  color: rgba(255, 255, 255, 0.85);
}

:deep(.el-input__wrapper) {
  background: rgba(28, 32, 58, 0.6) !important;
  border-radius: 8px;
  box-shadow: 0 1.6px 8px rgba(50, 80, 200, 0.07);
  border: 1px solid #22386736;
}

:deep(.el-input input) {
  color: #dde6ff;
  caret-color: #66aaff;
}

:deep(.el-select__input) {
  color: #dde6ff;
}

:deep(.el-select__placeholder) {
  color: rgba(255, 255, 255, 0.55);
}

:deep(.el-option) {
  color: rgba(255, 255, 255, 0.9);
}

:deep(.el-select-dropdown) {
  background: rgba(21, 21, 30, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.12);
}

:deep(.el-option:hover) {
  background: rgba(255, 255, 255, 0.05);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>

