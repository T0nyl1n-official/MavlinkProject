<template>
  <MainLayout v-if="showLayout" />
  <router-view v-else />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import MainLayout from './components/layout/MainLayout.vue'

const route = useRoute()
const authStore = useAuthStore()

const isLoginRoute = computed(() => route.path === '/login')
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
  background: #0A0E27;
  font-family: 'Microsoft YaHei', sans-serif;
}
</style>