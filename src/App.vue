<template>
  <!-- 科技网格线背景层 -->
  <div class="tech-grid"></div>

  <!-- 顶部流动光带 -->
  <div class="top-light-bar"></div>

  <!-- 扫描线效果 -->
  <div class="scanline-overlay"></div>

  <MainLayout v-if="showLayout" />
  <router-view v-else v-slot="{ Component }">
    <transition name="page-fade" mode="out-in">
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import MainLayout from './components/layout/MainLayout.vue'

const route = useRoute()
const authStore = useAuthStore()

const isLoginRoute = computed(() => route.path === '/login' || route.path === '/register')
const isLoggedIn = computed(() => !!authStore.token || !!localStorage.getItem('token'))

// 根据路由与登录状态决定是否展示主布局
const showLayout = computed(() => !isLoginRoute.value && isLoggedIn.value)
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  width: 100%;
  height: 100%;
}

body {
  background: var(--bg-body);
  font-family: var(--font-tech);
}

/* 页面切换动画 */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: all 0.4s ease-out;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(30px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
