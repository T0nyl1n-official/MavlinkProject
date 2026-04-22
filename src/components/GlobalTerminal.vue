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
    const response = await fetch('/terminal/message', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({ cmd, objects, args })
    })
    const result = await response.json()
    history.value.push({ cmd: rawCommand, result: `✓ ${JSON.stringify(result.message, null, 2)}` })
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

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})

// 暴露方法
defineExpose({
  open,
  close
})
</script>

<style scoped>
.global-terminal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
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
  background: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
  animation: slideIn 0.3s ease;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #252526;
  border-bottom: 1px solid #333;
}

.terminal-header h3 {
  color: #fff;
  margin: 0;
  font-size: 16px;
  font-weight: 500;
}

.close-btn {
  background: none;
  border: none;
  color: #888;
  font-size: 20px;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-btn:hover {
  background: #333;
  color: #fff;
}

.terminal-body {
  display: flex;
  flex-direction: column;
  height: 400px;
}

.terminal-output {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  font-family: 'Courier New', monospace;
  font-size: 14px;
}

.terminal-line {
  margin-bottom: 12px;
}

.command-prompt {
  color: #569cd6;
  margin-right: 8px;
}

.command-text {
  color: #d4d4d4;
}

.command-result {
  color: #4ec9b0;
  margin-left: 20px;
  margin-top: 4px;
  white-space: pre-wrap;
}

.loading {
  color: #569cd6;
  margin-top: 8px;
}

.terminal-input {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: #252526;
  border-top: 1px solid #333;
  gap: 12px;
}

.input-prompt {
  color: #569cd6;
  font-size: 14px;
  font-family: 'Courier New', monospace;
}

.terminal-input .el-input {
  flex: 1;
  background: #1e1e1e;
  border: 1px solid #333;
  border-radius: 4px;
}

.terminal-input .el-input__wrapper {
  background: #1e1e1e;
}

.terminal-input .el-input__input {
  color: #d4d4d4;
  font-family: 'Courier New', monospace;
}

.terminal-input .el-button {
  background: #0e639c;
  border-color: #1177bb;
}

.terminal-input .el-button:hover {
  background: #1177bb;
  border-color: #1177bb;
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
  background: #1e1e1e;
}

.terminal-output::-webkit-scrollbar-thumb {
  background: #333;
  border-radius: 4px;
}

.terminal-output::-webkit-scrollbar-thumb:hover {
  background: #444;
}
</style>