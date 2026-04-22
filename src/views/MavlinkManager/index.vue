<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage, ElDialog, ElForm, ElFormItem, ElInput, ElSelect, ElOption, ElButton, ElTable, ElTableColumn, ElTag, ElTabs, ElTabPane } from 'element-plus'
import { useMavlinkStore } from '@/stores/mavlink'
import type { MavlinkCommandParams } from '@/types/mavlink'

const mavlinkStore = useMavlinkStore()
const { connections, loading } = storeToRefs(mavlinkStore)

// 对话框状态
const connectDialogVisible = ref(false)
const sendCommandDialogVisible = ref(false)

// 表单数据
const connectForm = ref({
  version: 'v2',
  ip: '127.0.0.1',
  port: 14550,
  sysid: 1,
  compid: 50
})

const sendCommandForm = ref({
  connectionId: '',
  command: 'MAV_CMD_NAV_TAKEOFF',
  params: [0, 0, 0, 0, 0, 0, 10]
})

// 表单规则
const connectRules = {
  ip: [{ required: true, message: '请输入IP地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  sysid: [{ required: true, message: '请输入系统ID', trigger: 'blur' }],
  compid: [{ required: true, message: '请输入组件ID', trigger: 'blur' }]
}

const sendCommandRules = {
  connectionId: [{ required: true, message: '请选择连接', trigger: 'change' }],
  command: [{ required: true, message: '请选择命令', trigger: 'change' }]
}

// 生命周期
onMounted(() => {
  mavlinkStore.fetchConnections()
})

// 方法
async function handleConnect() {
  try {
    const success = await mavlinkStore.connect(connectForm.value)
    if (success) {
      ElMessage.success('连接成功')
      connectDialogVisible.value = false
      // 重置表单
      connectForm.value = {
        version: 'v2',
        ip: '127.0.0.1',
        port: 14550,
        sysid: 1,
        compid: 50
      }
      await mavlinkStore.fetchConnections()
    } else {
      ElMessage.error('连接失败')
    }
  } catch (error) {
    ElMessage.error('连接失败')
  }
}

async function handleDisconnect(connectionId: string) {
  try {
    const success = await mavlinkStore.disconnect(connectionId)
    if (success) {
      ElMessage.success('断开连接成功')
      await mavlinkStore.fetchConnections()
    } else {
      ElMessage.error('断开连接失败')
    }
  } catch (error) {
    ElMessage.error('断开连接失败')
  }
}

async function handleSendCommand() {
  try {
    const success = await mavlinkStore.sendCommand(sendCommandForm.value as MavlinkCommandParams)
    if (success) {
      ElMessage.success('发送命令成功')
      sendCommandDialogVisible.value = false
    } else {
      ElMessage.error('发送命令失败')
    }
  } catch (error) {
    ElMessage.error('发送命令失败')
  }
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h2 class="title">MAVLink 管理</h2>
      <div class="actions">
        <el-button size="small" @click="connectDialogVisible = true" type="primary">创建连接</el-button>
        <el-button size="small" @click="sendCommandDialogVisible = true" type="success">发送命令</el-button>
      </div>
    </div>

    <el-tabs>
      <el-tab-pane label="连接管理">
        <el-table :data="connections" style="width: 100%" row-key="id" v-loading="loading">
          <el-table-column prop="id" label="连接ID" />
          <el-table-column prop="version" label="版本" />
          <el-table-column prop="ip" label="IP" />
          <el-table-column prop="port" label="端口" />
          <el-table-column prop="sysid" label="系统ID" />
          <el-table-column prop="compid" label="组件ID" />
          <el-table-column label="状态">
            <template #default="{ row }">
              <el-tag :type="row.connected ? 'success' : 'info'">
                {{ row.connected ? '已连接' : '未连接' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template #default="{ row }">
              <el-button size="small" v-if="row.connected" @click="handleDisconnect(row.id)" type="danger">断开</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="命令列表">
        <div class="command-list">
          <el-card v-for="command in commandList" :key="command.id" class="command-card">
            <template #header>
              <div class="command-header">
                <span>{{ command.name }}</span>
                <el-tag size="small">{{ command.id }}</el-tag>
              </div>
            </template>
            <div class="command-description">{{ command.description }}</div>
            <div class="command-params">
              <h4>参数:</h4>
              <ul>
                <li v-for="(param, index) in command.params" :key="index">
                  {{ param }}
                </li>
              </ul>
            </div>
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 创建连接对话框 -->
    <el-dialog
      v-model="connectDialogVisible"
      title="创建 MAVLink 连接"
      width="500px"
    >
      <el-form :model="connectForm" :rules="connectRules" ref="connectFormRef">
        <el-form-item label="版本" prop="version">
          <el-select v-model="connectForm.version" placeholder="请选择版本">
            <el-option label="v1" value="v1" />
            <el-option label="v2" value="v2" />
          </el-select>
        </el-form-item>
        <el-form-item label="IP地址" prop="ip">
          <el-input v-model="connectForm.ip" placeholder="请输入IP地址" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input v-model.number="connectForm.port" type="number" placeholder="请输入端口" />
        </el-form-item>
        <el-form-item label="系统ID" prop="sysid">
          <el-input v-model.number="connectForm.sysid" type="number" placeholder="请输入系统ID" />
        </el-form-item>
        <el-form-item label="组件ID" prop="compid">
          <el-input v-model.number="connectForm.compid" type="number" placeholder="请输入组件ID" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="connectDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleConnect">连接</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 发送命令对话框 -->
    <el-dialog
      v-model="sendCommandDialogVisible"
      title="发送 MAVLink 命令"
      width="500px"
    >
      <el-form :model="sendCommandForm" :rules="sendCommandRules" ref="sendCommandFormRef">
        <el-form-item label="连接" prop="connectionId">
          <el-select v-model="sendCommandForm.connectionId" placeholder="请选择连接">
            <el-option v-for="conn in connections" :key="conn.id" :label="conn.id" :value="conn.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="命令" prop="command">
          <el-select v-model="sendCommandForm.command" placeholder="请选择命令">
            <el-option label="起飞" value="MAV_CMD_NAV_TAKEOFF" />
            <el-option label="降落" value="MAV_CMD_NAV_LAND" />
            <el-option label="移动" value="MAV_CMD_NAV_WAYPOINT" />
            <el-option label="返航" value="MAV_CMD_NAV_RETURN_TO_LAUNCH" />
            <el-option label="悬停" value="MAV_CMD_NAV_LOITER_TIME" />
          </el-select>
        </el-form-item>
        <el-form-item label="参数">
          <div class="params-grid">
            <el-input v-for="(_, index) in sendCommandForm.params" :key="index" v-model.number="sendCommandForm.params[index]" type="number" :placeholder="`参数 ${index + 1}`" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="sendCommandDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSendCommand">发送</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts">
export default {
  data() {
    return {
      commandList: [
        {
          id: 'MAV_CMD_NAV_TAKEOFF',
          name: '起飞',
          description: '无人机起飞到指定高度',
          params: [
            '空速 (m/s)',
            '空速类型 (0=地面速度, 1=空速)',
            '起飞类型 (0=垂直起飞, 1=滑行起飞)',
            'yaw角度 (度)',
            '纬度 (度)',
            '经度 (度)',
            '高度 (米)'
          ]
        },
        {
          id: 'MAV_CMD_NAV_LAND',
          name: '降落',
          description: '无人机降落到指定位置',
          params: [
            '空速 (m/s)',
            '空速类型 (0=地面速度, 1=空速)',
            '降落类型 (0=垂直降落, 1=滑行降落)',
            'yaw角度 (度)',
            '纬度 (度)',
            '经度 (度)',
            '高度 (米)'
          ]
        },
        {
          id: 'MAV_CMD_NAV_WAYPOINT',
          name: '移动到航点',
          description: '无人机移动到指定航点',
          params: [
            '空速 (m/s)',
            '空速类型 (0=地面速度, 1=空速)',
            '导航框架 (0=全局坐标系, 3=本地坐标系)',
            'yaw角度 (度)',
            '纬度 (度)',
            '经度 (度)',
            '高度 (米)'
          ]
        },
        {
          id: 'MAV_CMD_NAV_RETURN_TO_LAUNCH',
          name: '返航',
          description: '无人机返回起飞点',
          params: [
            '空速 (m/s)',
            '空速类型 (0=地面速度, 1=空速)',
            '返航类型 (0=直接返回, 1=从当前高度返回)',
            'yaw角度 (度)',
            '空 (保留)',
            '空 (保留)',
            '高度 (米)'
          ]
        },
        {
          id: 'MAV_CMD_NAV_LOITER_TIME',
          name: '悬停指定时间',
          description: '无人机在当前位置悬停指定时间',
          params: [
            '悬停时间 (秒)',
            '空速 (m/s)',
            '空速类型 (0=地面速度, 1=空速)',
            'yaw角度 (度)',
            '空 (保留)',
            '空 (保留)',
            '高度 (米)'
          ]
        }
      ]
    }
  }
}
</script>

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
  gap: 10px;
  flex-wrap: wrap;
}

.command-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.command-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s;
}

.command-card:hover {
  border-color: rgba(102, 170, 255, 0.3);
  box-shadow: 0 4px 12px rgba(102, 170, 255, 0.1);
}

.command-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.command-description {
  margin: 10px 0;
  color: rgba(255, 255, 255, 0.75);
  font-size: 14px;
}

.command-params h4 {
  margin: 10px 0;
  font-size: 14px;
  font-weight: 600;
}

.command-params ul {
  margin: 0;
  padding-left: 20px;
}

.command-params li {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.65);
  margin: 4px 0;
}

.params-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
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

:deep(.el-tabs__header) {
  margin-bottom: 16px;
}

:deep(.el-tabs__tab) {
  color: rgba(255, 255, 255, 0.75);
}

:deep(.el-tabs__tab.is-active) {
  color: #66aaff;
}

:deep(.el-tabs__active-bar) {
  background-color: #66aaff;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>