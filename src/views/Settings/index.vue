<template>
  <div style="padding:24px">
    <h1 class="gradient-title">⚙️ 服务器设置</h1>
    <div style="display:grid; grid-template-columns:repeat(auto-fit,minmax(350px,1fr)); gap:24px">
      <div class="soft-card">
        <h3>📡 连接设置</h3>
        <el-form label-width="100px">
          <el-form-item label="UDP端口"><el-input-number v-model="s.udp_port" :min="1" :max="65535"/></el-form-item>
          <el-form-item label="连接超时"><el-input-number v-model="s.timeout" :min="1"/> 秒</el-form-item>
          <el-form-item label="自动重连"><el-switch v-model="s.auto_reconnect"/></el-form-item>
        </el-form>
      </div>
      <div class="soft-card">
        <h3>📊 日志设置</h3>
        <el-form label-width="100px">
          <el-form-item label="日志级别"><el-select v-model="s.log_level"><el-option label="INFO" value="info"/><el-option label="DEBUG" value="debug"/></el-select></el-form-item>
          <el-form-item label="日志保留"><el-input-number v-model="s.log_retention" :min="1"/> 天</el-form-item>
        </el-form>
      </div>
      <div class="soft-card">
        <h3>🔐 安全设置</h3>
        <el-form label-width="100px">
          <el-form-item label="最大连接数"><el-input-number v-model="s.max_connections" :min="1"/></el-form-item>
          <el-form-item label="IP白名单"><el-input v-model="s.ip_whitelist" placeholder="多个IP用逗号分隔"/></el-form-item>
        </el-form>
      </div>
      <div class="soft-card">
        <h3>🔌 联调模式</h3>
        <el-form label-width="120px">
          <el-form-item label="真实后端联调模式">
            <el-switch v-model="useRealApi" @change="handleModeChange" />
          </el-form-item>
          <el-form-item label="当前状态">
            <el-tag :type="useRealApi ? 'success' : 'info'">
              {{ useRealApi ? '真实后端' : 'Mock 模式' }}
            </el-tag>
          </el-form-item>
          <el-form-item label="说明">
            <span style="color: var(--text-secondary); font-size: 12px;">
              开启后将调用真实后端 API，需后端服务运行中
            </span>
          </el-form-item>
        </el-form>
      </div>
    </div>
    <div style="text-align:center; margin-top:32px">
      <el-button type="primary" @click="save" :loading="saving">💾 保存</el-button>
      <el-button @click="reset">🔄 重置</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { config } from '@/utils/mockService'

const saving = ref(false)
const s = ref({ udp_port:14550, timeout:30, auto_reconnect:true, log_level:'info', log_retention:7, max_connections:100, ip_whitelist:'' })
const useRealApi = ref(config.USE_REAL_API)

const save = async () => { saving.value=true; try{ await fetch('/api/settings',{method:'POST',body:JSON.stringify(s.value),headers:{'Content-Type':'application/json'}}); ElMessage.success('已保存') }catch(e){ ElMessage.error('失败') } finally{ saving.value=false } }
const reset = () => { s.value = { udp_port:14550, timeout:30, auto_reconnect:true, log_level:'info', log_retention:7, max_connections:100, ip_whitelist:'' }; ElMessage.info('已重置') }
const handleModeChange = async (value: boolean) => {
  try {
    await ElMessageBox.confirm(
      `切换到 ${value ? '真实后端' : 'Mock'} 模式后，需要刷新页面才能生效。是否立即刷新？`,
      '提示',
      {
        confirmButtonText: '立即刷新',
        cancelButtonText: '稍后刷新',
        type: 'warning'
      }
    )
    config.USE_REAL_API = value
    localStorage.setItem('useRealApi', String(value))
    window.location.reload()
  } catch {
    useRealApi.value = config.USE_REAL_API
  }
}
onMounted(async () => {
  const saved = localStorage.getItem('useRealApi')
  if (saved) {
    config.USE_REAL_API = saved === 'true'
    useRealApi.value = config.USE_REAL_API
  }
  try{ const res = await fetch('/api/settings'); const data = await res.json(); s.value = data.data }catch(e){}
})
</script>

<style scoped>
.soft-card {
  background: var(--bg-card);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  padding: 24px;
  transition: all 0.3s ease;
}

.soft-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-hover);
}

.gradient-title {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 32px;
  text-align: center;
  letter-spacing: 2px;
  background: linear-gradient(135deg, var(--primary), var(--primary-light));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
</style>