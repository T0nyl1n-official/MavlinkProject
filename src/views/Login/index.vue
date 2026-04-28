<template>
  <div class="login-bg">
    <div class="login-wrapper">
      <!-- 无人机插画 -->
      <div class="login-illustration">
        <img src="/logo.svg" alt="深地哨兵" style="width: 200px; height: 200px;" />
      </div>
      
      <!-- 登录表单 -->
      <div class="login-container">
        <el-form
          :model="credentials"
          :rules="formRules"
          class="login-form"
          @submit.prevent="handleLogin"
        >
          <h2 class="login-title gradient-title">登录</h2>
          <el-form-item prop="email">
            <el-input
              v-model="credentials.email"
              prefix-icon="el-icon-user"
              placeholder="邮箱"
              autocomplete="email"
              clearable
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="credentials.password"
              prefix-icon="el-icon-lock"
              placeholder="密码"
              autocomplete="current-password"
              show-password
              clearable
              type="password"
            />
          </el-form-item>
          <el-form-item>
            <el-button
              :loading="loading"
              type="primary"
              class="login-btn"
              native-type="submit"
              round
            >
              <span v-if="!loading">登录</span>
            </el-button>
          </el-form-item>
        </el-form>
        <div class="login-footer">
          <span>还没有账号？</span>
          <router-link to="/register" class="link">注册</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const loading = ref(false)
const authStore = useAuthStore()

const credentials = reactive({
  email: '',
  password: ''
})

const formRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度应为6-20个字符', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  loading.value = true
  try {
    const response = await authStore.login({
      email: credentials.email,
      password: credentials.password
    })
    
    if (response.success && response.data?.token) {
      ElMessage.success(response.data.message || response.message || '欢迎回来')
      await router.push('/chain-create')
    } else {
      ElMessage.error(response.data?.message || response.message || '登录没有成功')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '登录失败，请稍后再试')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-bg {
  min-height: 100vh;
  width: 100vw;
  background: var(--bg-body);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.login-wrapper {
  display: flex;
  align-items: center;
  gap: 48px;
  max-width: 800px;
  width: 100%;
  animation: floatCard 1.2s cubic-bezier(.4,0,.2,1);
}

@keyframes floatCard {
  from {
    opacity: 0;
    transform: translateY(48px) scale(0.96);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.login-illustration {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}

.login-container {
  flex: 1;
  background: var(--bg-card);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  padding: 48px 38px 36px 38px;
  min-width: 320px;
  max-width: 400px;
  width: 100%;
  transition: all 0.3s ease;
}

.login-container:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.login-title {
  text-align: center;
  letter-spacing: 2px;
  margin-bottom: 32px;
  font-size: 2rem;
  font-weight: 600;
  user-select: none;
}

.el-form-item {
  margin-bottom: 26px;
}

.el-input {
  font-size: 1rem;
}

.login-btn {
  width: 100%;
  height: 42px;
  font-size: 1.1rem;
  transition: all 0.32s cubic-bezier(.17,.67,.37,1.34);
  letter-spacing: 1px;
}

.login-btn:hover {
  transform: translateY(-2px);
}

.login-btn:active {
  transform: scale(0.98);
}

.el-form-item__error {
  color: #ff4488;
  font-size: 13px;
  margin-top: 2px;
}

.login-footer {
  margin-top: 20px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 14px;
}

.link {
  color: var(--primary);
  text-decoration: none;
  margin-left: 8px;
  transition: color 0.3s;
}

.link:hover {
  color: var(--primary-light);
  text-decoration: underline;
}

@media (max-width: 768px) {
  .login-wrapper {
    flex-direction: column;
    gap: 32px;
  }
  
  .login-illustration {
    order: 1;
  }
  
  .login-container {
    order: 2;
  }
}
</style>