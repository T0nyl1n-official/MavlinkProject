<template>
  <div v-if="visible" class="global-terminal-overlay" @click="close">
    <div class="global-terminal" @click.stop>
      <div class="terminal-header">
        <h3>💻 终端</h3>
        <button class="close-btn" @click="close">×</button>
      </div>
      <div class="terminal-body">
        <div ref="outputRef" class="terminal-output">
          <div v-for="(item, i) in history" :key="i" class="terminal-line">
            <span class="command-prompt">$</span>
            <span class="command-text">{{ item.cmd }}</span>
            <div class="command-result">{{ item.result }}</div>
          </div>
          <div v-if="loading" class="loading">执行中...</div>
        </div>
        <div class="terminal-input">
          <span class="input-prompt">$</span>
          <el-input
            v-model="command"
            placeholder="输入命令..."
            @keyup.enter="sendCommand"
            @keyup.esc="close"
            :disabled="loading"
            ref="inputRef"
            autofocus
          />
          <el-button @click="sendCommand" :loading="loading">发送</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { config } from '@/utils/mockService'
import { mockService } from '@/utils/mockService'

onMounted(() => {
  const saved = localStorage.getItem('useRealApi')
  if (saved) {
    config.USE_REAL_API = saved === 'true'
  }
  window.addEventListener('keydown', handleKeyDown)
})

const visible = ref(false)
const command = ref('')
const loading = ref(false)
const history = ref<{ cmd: string, result: string }[]>([])
const outputRef = ref<HTMLElement>()
const inputRef = ref<HTMLElement>()

// 发送命令到后端终端接口
const sendCommand = async () => {
  if (!command.value.trim()) return
  
  const rawCommand = command.value.trim()
  command.value = ''
  loading.value = true
  
  // 解析命令（支持 help/whoami/ls/server/backend 等）
  const parts = rawCommand.split(/\s+/)
  const cmd = parts[0]
  const objects = parts.slice(1).filter(p => !p.startsWith('--'))
  const args: Record<string, any> = {}
  
  // 解析 --key value 格式的参数
  for (let i = 1; i < parts.length; i++) {
    if (parts[i].startsWith('--')) {
      const key = parts[i].slice(2)
      const value = parts[i + 1] && !parts[i + 1].startsWith('--') ? parts[i + 1] : true
      args[key] = isNaN(Number(value)) ? value : Number(value)
      if (value !== true) i++
    }
  }
  
  try {
    if (config.USE_REAL_API) {
      const token = localStorage.getItem('token')
      if (!token) {
        throw new Error('未检测到登录 token，请先重新登录')
      }
      const response = await fetch('/terminal/message', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ cmd, objects, args })
      })
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || `请求失败(${response.status})`)
      }
      const result = await response.json()
      history.value.push({ cmd: rawCommand, result: `✓ ${JSON.stringify(result.message, null, 2)}` })
    } else {
      const result = await mockService.executeTerminalCommand(rawCommand)
      history.value.push({ cmd: rawCommand, result: `✓ ${result}` })
    }
  } catch (error: any) {
    history.value.push({ cmd: rawCommand, result: `✗ 命令执行失败: ${error.message}` })
  } finally {
    loading.value = false
    // 滚动到最新输出
    nextTick(() => {
      if (outputRef.value) {
        outputRef.value.scrollTop = outputRef.value.scrollHeight
      }
    })
  }
}

const open = () => {
  visible.value = true
  nextTick(() => {
    if (inputRef.value) {
      (inputRef.value as any).focus()
    }
  })
}

const close = () => {
  visible.value = false
  command.value = ''
}

// 键盘事件监听
const handleKeyDown = (e: KeyboardEvent) => {
  // 按 / 键打开终端（确保不是在输入框中）
  const target = e.target
  if (
    e.key === '/' &&
    target instanceof HTMLElement &&
    !/INPUT|TEXTAREA/i.test(target.tagName)
  ) {
    e.preventDefault()
    open()
  }
  // 按 ESC 键关闭终端
  if (e.key === 'Escape' && visible.value) {
    close()
  }
}

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})

// 添加日志
const addLog = (message: string) => {
  history.value.push({ cmd: '[SYSTEM]', result: message })
  // 滚动到最新输出
  nextTick(() => {
    if (outputRef.value) {
      outputRef.value.scrollTop = outputRef.value.scrollHeight
    }
  })
}

// 暴露方法
defineExpose({
  open,
  close,
  addLog
})
</script>

<style scoped>
.global-terminal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  animation: fadeIn 0.3s ease;
}

.global-terminal {
  width: 800px;
  max-width: 90vw;
  max-height: 80vh;
  background: #0d1117;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 10px 40px rgba(42, 82, 152, 0.5);
  border: 1px solid #2a5298;
  animation: slideIn 0.3s ease;
  position: relative;
}

/* 扫描线动画 */
.global-terminal::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    transparent 50%,
    rgba(42, 82, 152, 0.03) 50%
  );
  background-size: 100% 4px;
  pointer-events: none;
  z-index: 10;
}

/* 扫描线移动 */
.global-terminal::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, #6c9bd1, transparent);
  animation: scanLine 3s linear infinite;
  pointer-events: none;
  z-index: 11;
}

@keyframes scanLine {
  0% { top: 0; opacity: 0.8; }
  50% { opacity: 0.4; }
  100% { top: 100%; opacity: 0.8; }
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #121826;
  border-bottom: 1px solid #1e3c72;
  position: relative;
  z-index: 12;
}

.terminal-header h3 {
  color: #6c9bd1;
  margin: 0;
  font-size: 16px;
  font-weight: 500;
  font-family: 'Courier New', monospace;
}

.close-btn {
  background: none;
  border: none;
  color: #9cb8d8;
  font-size: 20px;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: #1e3c72;
  color: #fff;
}

.terminal-body {
  display: flex;
  flex-direction: column;
  height: 400px;
  position: relative;
  z-index: 12;
}

.terminal-output {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  position: relative;
}

.terminal-line {
  margin-bottom: 12px;
}

.command-prompt {
  color: #6c9bd1;
  margin-right: 8px;
}

.command-text {
  color: #e6f0ff;
}

.command-result {
  color: #4ec9b0;
  margin-left: 20px;
  margin-top: 4px;
  white-space: pre-wrap;
}

.loading {
  color: #6c9bd1;
  margin-top: 8px;
  position: relative;
}

/* 加载动画 */
.loading::after {
  content: '';
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #6c9bd1;
  margin-left: 8px;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  50% { opacity: 0; }
}

.terminal-input {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: #121826;
  border-top: 1px solid #1e3c72;
  gap: 12px;
  position: relative;
  z-index: 12;
}

.input-prompt {
  color: #6c9bd1;
  font-size: 14px;
  font-family: 'Courier New', monospace;
}

/* 光标闪烁效果 */
.input-prompt::after {
  content: '▊';
  color: #6c9bd1;
  animation: cursorBlink 1s step-end infinite;
  margin-left: 4px;
}

@keyframes cursorBlink {
  50% { opacity: 0; }
}

.terminal-input .el-input {
  flex: 1;
  background: #0d1117;
  border: 1px solid #1e3c72;
  border-radius: 4px;
}

.terminal-input .el-input__wrapper {
  background: #0d1117;
  box-shadow: none;
}

.terminal-input .el-input__input {
  color: #e6f0ff;
  font-family: 'Courier New', monospace;
}

.terminal-input .el-button {
  background: linear-gradient(135deg, #1e3c72, #2a5298);
  border-color: #2a5298;
  transition: all 0.2s ease;
}

.terminal-input .el-button:hover {
  background: linear-gradient(135deg, #2a5298, #1e3c72);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(42, 82, 152, 0.4);
}

.terminal-input .el-button:active {
  transform: scale(0.98);
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* 滚动条样式 */
.terminal-output::-webkit-scrollbar {
  width: 8px;
}

.terminal-output::-webkit-scrollbar-track {
  background: #0d1117;
}

.terminal-output::-webkit-scrollbar-thumb {
  background: #2a5298;
  border-radius: 4px;
}

.terminal-output::-webkit-scrollbar-thumb:hover {
  background: #1e3c72;
}
</style>