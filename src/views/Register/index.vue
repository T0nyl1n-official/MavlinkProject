<template>
  <div class="register-bg">
    <div class="register-wrapper">
      <!-- 火箭插画 -->
      <div class="register-illustration">
        <svg width="200" height="200" viewBox="0 0 200 200" fill="none">
          <circle cx="100" cy="100" r="80" fill="#E8DDFF" opacity="0.5"/>
          <path d="M100 30L120 70H80L100 30Z" fill="#8C7CF0"/>
          <rect x="95" y="70" width="10" height="60" fill="#C6B9FF" rx="5"/>
          <circle cx="100" cy="85" r="8" fill="#FFD3B6"/>
          <path d="M90 130L100 160L110 130H90Z" fill="#FFE5B4"/>
          <circle cx="70" cy="50" r="3" fill="#C6B9FF"/>
          <circle cx="130" cy="60" r="2" fill="#C6B9FF"/>
          <circle cx="55" cy="80" r="2" fill="#C6B9FF"/>
        </svg>
      </div>
      
      <!-- 注册表单 -->
      <div class="register-container">
        <el-form
          ref="registerFormRef"
          :model="registerForm"
          :rules="registerRules"
          class="register-form"
          @submit.prevent="handleRegister"
        >
          <h2 class="register-title gradient-title">注册</h2>
          <el-form-item prop="username">
            <el-input
              v-model="registerForm.username"
              prefix-icon="el-icon-user"
              placeholder="用户名"
              autocomplete="username"
              clearable
            />
          </el-form-item>
          <el-form-item prop="email">
            <el-input
              v-model="registerForm.email"
              prefix-icon="el-icon-message"
              placeholder="邮箱"
              autocomplete="email"
              clearable
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="registerForm.password"
              prefix-icon="el-icon-lock"
              placeholder="密码"
              autocomplete="new-password"
              show-password
              clearable
              type="password"
            />
          </el-form-item>
          <el-form-item prop="confirmPassword">
            <el-input
              v-model="registerForm.confirmPassword"
              prefix-icon="el-icon-lock"
              placeholder="确认密码"
              autocomplete="new-password"
              show-password
              clearable
              type="password"
            />
          </el-form-item>
          <el-form-item>
            <el-button
              :loading="loading"
              type="primary"
              class="register-btn"
              @click="submitForm"
              round
            >
              <span v-if="!loading">注册</span>
            </el-button>
          </el-form-item>
        </el-form>
        <div class="register-footer">
          <span>已有账号？</span>
          <router-link to="/login" class="link">登录</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { registerApi } from '@/api/auth'

const router = useRouter()
const registerFormRef = ref()
const loading = ref(false)

const registerForm = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const registerRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度应为3-20个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度应为6-20个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: any) => {
        if (value !== registerForm.password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

function doRegister() {
  if (!registerFormRef.value) return

  registerFormRef.value.validate(async (valid: boolean) => {
    if (!valid) return

    loading.value = true
    try {
      const res = await registerApi({
        username: registerForm.username,
        email: registerForm.email,
        password: registerForm.password
      })
      if (res.success) {
        ElMessage.success('注册成功！')
        await router.push('/login')
      } else {
        ElMessage.error(res.message || '注册失败')
      }
    } catch (e: any) {
      ElMessage.error(e?.message || '注册失败')
    } finally {
      loading.value = false
    }
  })
}

function submitForm() {
  doRegister()
}

function handleRegister(e: Event) {
  e.preventDefault()
  doRegister()
}
</script>

<style scoped>
.register-bg {
  min-height: 100vh;
  width: 100vw;
  background: var(--bg-body);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.register-wrapper {
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

.register-illustration {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}

.register-container {
  flex: 1;
  background: var(--bg-card);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  padding: 48px 38px 36px 38px;
  min-width: 320px;
  max-width: 400px;
  width: 100%;
  transition: all 0.3s ease;
}

.register-container:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-hover);
}

.register-title {
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

.register-btn {
  width: 100%;
  height: 42px;
  font-size: 1.1rem;
  transition: all 0.32s cubic-bezier(.17,.67,.37,1.34);
  letter-spacing: 1px;
}

.register-btn:hover {
  transform: translateY(-2px);
}

.register-btn:active {
  transform: scale(0.98);
}

/* 修改 ElementPlus 表单校验提示颜色 */
.el-form-item__error {
  color: #ff4488;
  font-size: 13px;
  margin-top: 2px;
}

.register-footer {
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

/* 响应式设计 */
@media (max-width: 768px) {
  .register-wrapper {
    flex-direction: column;
    gap: 32px;
  }
  
  .register-illustration {
    order: 1;
  }
  
  .register-container {
    order: 2;
  }
}
</style>