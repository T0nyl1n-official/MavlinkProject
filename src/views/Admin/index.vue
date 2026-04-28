<template>
  <div class="admin-container">
    <h1 class="gradient-title">👥 用户管理</h1>

    <div class="admin-card">
      <div class="card-header">
        <h2>所有用户列表</h2>
        <el-button type="primary" @click="fetchUsers" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column prop="User_ID" label="用户ID" width="100" />
        <el-table-column prop="Username" label="用户名" />
        <el-table-column prop="Email" label="邮箱" />
        <el-table-column prop="Role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.Role === 'admin' ? 'danger' : 'info'">
              {{ row.Role === 'admin' ? '管理员' : '普通用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button
              v-if="row.Role !== 'admin'"
              type="warning"
              size="small"
              @click="handleSetAdmin(row)"
            >
              设为管理员
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(row)"
              :disabled="row.User_ID === currentUserId"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { getAllUsersApi, deleteUserApi, updateUserRoleApi } from '@/api/admin'
import { useAuthStore } from '@/stores/auth'

interface User {
  User_ID: number
  Username: string
  Email: string
  Role: string
}

const authStore = useAuthStore()
const users = ref<User[]>([])
const loading = ref(false)

const currentUserId = computed(() => {
  const user = authStore.userInfo as Record<string, unknown> | null
  if (!user) return undefined
  return Number(user.User_ID || user.user_id || 0)
})

const fetchUsers = async () => {
  loading.value = true
  try {
    const res = await getAllUsersApi()
    users.value = res.data?.users || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleSetAdmin = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要将用户「${user.Username}」设为管理员吗？`,
      '提示',
      { type: 'warning' }
    )
    await updateUserRoleApi(user.User_ID, 'admin')
    ElMessage.success('设置成功')
    fetchUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '设置失败')
    }
  }
}

const handleDelete = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户「${user.Username}」吗？此操作不可恢复！`,
      '警告',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteUserApi(user.User_ID)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.admin-container {
  padding: 24px;
  min-height: 100vh;
  background: var(--bg-body);
}

.admin-card {
  background: var(--bg-card);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  padding: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.card-header h2 {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--text-primary);
}
</style>
