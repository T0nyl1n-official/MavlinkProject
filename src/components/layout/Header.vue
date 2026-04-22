<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

async function handleLogout() {
  try {
    await authStore.logout()
  } catch (e: unknown) {
    // 尽量不影响跳转
    ElMessage.error('退出登录失败')
  } finally {
    router.push('/login')
  }
}
</script>

<template>
  <el-header class="app-header">
    <div class="header-left">
      <div class="header-title">无人机平台管理</div>
    </div>
    <div class="header-right">
      <el-button type="primary" link @click="handleLogout" class="logout-btn">退出登录</el-button>
    </div>
  </el-header>
</template>

<style scoped>
.app-header {
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 32px;
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
  border-bottom: 1px solid var(--border-color);
}

.header-title {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 18px;
  font-family: 'Plus Jakarta Sans', sans-serif;
}

.logout-btn {
  color: var(--primary) !important;
  font-weight: 500;
}

.logout-btn:hover {
  color: var(--primary-light) !important;
}
</style>

