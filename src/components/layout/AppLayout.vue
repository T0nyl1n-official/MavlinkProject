<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const activePath = computed(() => route.path)

async function handleLogout() {
  try {
    await authStore.logout()
  } catch (e: any) {
    ElMessage.error(e?.message || '退出登录失败')
  } finally {
    router.push('/login')
  }
}
</script>

<template>
  <el-container class="app-layout">
    <el-aside class="app-aside">
      <div class="app-logo">深地哨兵</div>

      <el-menu :default-active="activePath" router class="app-menu">
        <el-menu-item index="/dashboard">仪表盘</el-menu-item>
        <el-menu-item index="/chain">任务链</el-menu-item>
        <el-menu-item index="/board">板子管理</el-menu-item>
        <el-menu-item index="/monitor">实时监控</el-menu-item>
      </el-menu>
    </el-aside>

    <el-container class="app-main">
      <el-header class="app-header">
        <div class="header-left">
          <div class="header-title">无人机平台管理</div>
        </div>
        <div class="header-right">
          <el-button type="primary" link @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>

      <el-main class="app-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.app-layout {
  height: 100%;
  background: #0a0e27;
}

.app-aside {
  width: 220px;
  background: rgba(255, 255, 255, 0.05);
  border-right: 1px solid rgba(255, 255, 255, 0.08);
}

.app-logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #66aaff;
  font-weight: 700;
  letter-spacing: 1px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.app-menu {
  background: transparent;
  padding-top: 10px;
}

/* 简化菜单样式，使其适配暗色背景 */
:deep(.el-menu) {
  border-right: none;
}

:deep(.el-menu-item) {
  color: rgba(255, 255, 255, 0.85);
}

:deep(.el-menu-item.is-active) {
  background: rgba(102, 170, 255, 0.18) !important;
  color: #66aaff !important;
}

.app-main {
  height: 100%;
}

.app-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 18px;
  background: rgba(255, 255, 255, 0.04);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.header-title {
  color: rgba(255, 255, 255, 0.9);
  font-weight: 600;
}

.app-content {
  padding: 18px;
}
</style>

