<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const activePath = computed(() => route.path)
const isAdmin = computed(() => localStorage.getItem('role') === 'admin')
</script>

<template>
  <el-aside class="app-aside">
    <div class="app-logo">
      <img src="/logo.svg" alt="深地哨兵" class="logo-icon" />
    </div>

    <el-menu :default-active="activePath" router class="app-menu" mode="vertical">
      <el-menu-item index="/chain-manager">
        <template #icon>
          <span class="menu-icon">📋</span>
        </template>
        <span class="menu-text">任务链</span>
      </el-menu-item>
      <el-menu-item index="/dashboard">
        <template #icon>
          <span class="menu-icon">📊</span>
        </template>
        <span class="menu-text">仪表盘</span>
      </el-menu-item>
      <el-menu-item index="/board">
        <template #icon>
          <span class="menu-icon">📱</span>
        </template>
        <span class="menu-text">板子管理</span>
      </el-menu-item>
      <el-menu-item index="/mavlink">
        <template #icon>
          <span class="menu-icon">✈️</span>
        </template>
        <span class="menu-text">MAVLink控制</span>
      </el-menu-item>
      <el-menu-item index="/monitor">
        <template #icon>
          <span class="menu-icon">📡</span>
        </template>
        <span class="menu-text">监控</span>
      </el-menu-item>
      <el-menu-item index="/admin" v-if="isAdmin">
        <template #icon>
          <span class="menu-icon">👥</span>
        </template>
        <span class="menu-text">用户管理</span>
      </el-menu-item>
      <el-menu-item index="/settings">
        <template #icon>
          <span class="menu-icon">⚙️</span>
        </template>
        <span class="menu-text">设置</span>
      </el-menu-item>
      <el-menu-item index="/terminal">
        <template #icon>
          <span class="menu-icon">💻</span>
        </template>
        <span class="menu-text">终端</span>
      </el-menu-item>
    </el-menu>
  </el-aside>
</template>

<style scoped>
.app-aside {
  width: 220px;
  background: var(--bg-menu);
  box-shadow: var(--shadow-md);
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  padding: 0;
}

.app-logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 16px;
  width: 100%;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.logo-icon {
  width: 36px;
  height: 36px;
  object-fit: contain;
}

.app-menu {
  width: 100%;
  background: transparent;
  padding-top: 16px;
}

:deep(.el-menu) {
  border-right: none;
  background: transparent;
  height: 100%;
}

:deep(.el-menu-item) {
  color: rgba(255,255,255,0.7);
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  margin: 0;
  border-radius: 0;
  padding: 0 20px;
  transition: all 0.3s ease;
}

:deep(.el-menu-item:hover) {
  background: rgba(255,255,255,0.1) !important;
  color: white !important;
}

:deep(.el-menu-item.is-active) {
  background: var(--primary) !important;
  color: white !important;
  box-shadow: none;
}

.menu-icon {
  font-size: 18px;
  margin-right: 12px;
  margin-bottom: 0;
}

.menu-text {
  font-size: 14px;
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .app-aside {
    width: 60px;
  }
  
  .menu-text {
    display: none;
  }
  
  :deep(.el-menu-item) {
    padding: 0 12px;
    justify-content: center;
  }
  
  .menu-icon {
    margin-right: 0;
  }
  
  .logo-icon {
    margin-right: 0;
  }
}
</style>

