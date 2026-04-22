<script setup lang="ts">
import Sidebar from './Sidebar.vue'
import Header from './Header.vue'
</script>

<template>
  <div class="main-layout">
    <!-- 左侧导航 -->
    <Sidebar />
    
    <!-- 中间内容区 -->
    <div class="main-content">
      <Header />
      <div class="content-area">
        <router-view />
      </div>
    </div>
    
    <!-- 右侧信息栏 -->
    <div class="right-sidebar">
      <h3 class="sidebar-title">系统信息</h3>
      <div class="info-card">
        <div class="info-item">
          <span class="info-label">状态</span>
          <span class="info-value">运行中</span>
        </div>
        <div class="info-item">
          <span class="info-label">无人机数量</span>
          <span class="info-value">3</span>
        </div>
        <div class="info-item">
          <span class="info-label">任务链</span>
          <span class="info-value">2</span>
        </div>
        <div class="info-item">
          <span class="info-label">板子数量</span>
          <span class="info-value">5</span>
        </div>
      </div>
      
      <h3 class="sidebar-title">最近活动</h3>
      <div class="activity-card">
        <div class="activity-item">
          <span class="activity-time">2分钟前</span>
          <span class="activity-text">无人机 #1 起飞</span>
        </div>
        <div class="activity-item">
          <span class="activity-time">15分钟前</span>
          <span class="activity-text">任务链 #2 完成</span>
        </div>
        <div class="activity-item">
          <span class="activity-time">1小时前</span>
          <span class="activity-text">板子 #3 连接成功</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.main-layout {
  display: flex;
  min-height: 100vh;
  background: var(--bg-body);
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.content-area {
  flex: 1;
  padding: 32px;
  background: var(--bg-body);
  overflow-y: auto;
}

.right-sidebar {
  width: 300px;
  background: var(--bg-card);
  box-shadow: var(--shadow-md);
  border-radius: 32px 0 0 32px;
  padding: 24px;
  flex-shrink: 0;
  overflow-y: auto;
}

.sidebar-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
  color: var(--text-primary);
}

.info-card {
  background: var(--bg-body);
  border-radius: var(--radius-sm);
  padding: 16px;
  margin-bottom: 24px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}

.info-label {
  color: var(--text-secondary);
  font-size: 14px;
}

.info-value {
  color: var(--text-primary);
  font-weight: 500;
  font-size: 14px;
}

.activity-card {
  background: var(--bg-body);
  border-radius: var(--radius-sm);
  padding: 16px;
}

.activity-item {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(140, 124, 240, 0.1);
}

.activity-item:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.activity-time {
  display: block;
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.activity-text {
  font-size: 14px;
  color: var(--text-primary);
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .right-sidebar {
    display: none;
  }
}

@media (max-width: 768px) {
  .content-area {
    padding: 16px;
  }
}
</style>

