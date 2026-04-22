<template>
  <div style="padding:24px">
    <h1 class="gradient-title">💻 终端控制台</h1>
    <p style="color:var(--text-secondary)">⚠️ 后端开发中，当前为模拟模式</p>
    <div style="background:#1e1e1e; border-radius:16px; overflow:hidden">
      <div ref="outRef" style="height:400px; overflow-y:auto; padding:16px; font-family:monospace">
        <div v-for="(item,i) in history" :key="i" style="margin-bottom:16px">
          <div style="color:#4ec9b0"><span style="color:#569cd6">$</span> {{ item.cmd }}</div>
          <div style="color:#d4d4d4; padding-left:20px; border-left:2px solid #4ec9b0" v-html="item.out.replace(/\n/g,'<br>')"></div>
        </div>
        <div v-if="loading"><span style="color:#569cd6">></span> 执行中...</div>
      </div>
      <div style="display:flex; gap:12px; padding:16px; border-top:1px solid #333; background:#252526">
        <span style="color:#569cd6; font-size:18px">$</span>
        <el-input v-model="cmd" placeholder="输入指令..." @keyup.enter="send" :disabled="loading"/>
        <el-button @click="send" :loading="loading">发送</el-button>
        <el-button @click="history=[]">清屏</el-button>
      </div>
      <div style="display:flex; gap:8px; padding:12px 16px; border-top:1px solid #333; background:#252526">
        <span style="color:#888">快捷：</span>
        <el-button size="small" @click="quick('get status')">get status</el-button>
        <el-button size="small" @click="quick('get settings')">get settings</el-button>
        <el-button size="small" @click="quick('list chains')">list chains</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { mockService } from '@/utils/mockService'
const cmd = ref('')
const loading = ref(false)
const history = ref<{cmd:string,out:string}[]>([])
const outRef = ref()
const send = async () => {
  if(!cmd.value.trim()) return
  const command = cmd.value.trim()
  cmd.value = ''
  loading.value = true
  try {
    const response = await mockService.executeTerminalCommand(command)
    history.value.push({ cmd: command, out: response })
  } catch (error) {
    history.value.push({ cmd: command, out: `Error: ${error}` })
  } finally {
    loading.value = false
    setTimeout(() => { if(outRef.value) outRef.value.scrollTop = outRef.value.scrollHeight }, 100)
  }
}
const quick = (c:string) => { cmd.value = c; send() }
</script>

<style scoped>
.gradient-title {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 16px;
  text-align: center;
  letter-spacing: 2px;
  background: linear-gradient(135deg, var(--primary), var(--primary-light));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
</style>