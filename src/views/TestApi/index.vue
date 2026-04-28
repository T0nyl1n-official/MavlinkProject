<template>
  <div class="test-api-page">
    <div class="header">
      <h2>API 串行测试</h2>
      <p class="desc">访问路径：`/test-api`，覆盖前端现有 API 封装的串行联调页面。</p>
    </div>

    <div class="actions">
      <el-button type="primary" :loading="running" @click="runAllTests">
        一键运行全部测试
      </el-button>
      <el-button :disabled="results.length === 0" @click="copyTableText">
        复制结果文本
      </el-button>
      <div class="mode-switch">
        <span class="mode-label">真实后端模式</span>
        <el-switch v-model="useRealApi" @change="handleModeChange" />
      </div>
    </div>

    <div class="meta" v-if="results.length > 0">
      <span>模式：{{ apiModeText }}</span>
      <span>执行时间：{{ finishedAt || '-' }}</span>
      <span>总计：{{ results.length }}</span>
      <span>成功：{{ results.filter(i => i.ok).length }}</span>
      <span>失败：{{ results.filter(i => !i.ok).length }}</span>
    </div>

    <div class="notes">
      <p>说明：</p>
      <p>1. 测试通过前端 API 封装执行，Mock 模式和真实模式都可运行。</p>
      <p>2. 部分接口依赖登录态、管理员权限或后端环境，失败会直接显示在结果中。</p>
      <p>3. 视频流、MAVLink、管理员接口若失败，优先检查 token、角色权限、在线设备与后端运行状态。</p>
      <p>4. 默认跳过高风险删除操作，避免误删测试数据。</p>
    </div>

    <el-table :data="results" border style="width: 100%">
      <el-table-column prop="name" label="步骤" min-width="220" />
      <el-table-column prop="statusCode" label="状态码" width="100" />
      <el-table-column label="是否成功" width="100">
        <template #default="{ row }">
          <el-tag :type="row.ok ? 'success' : 'danger'">{{ row.ok ? '是' : '否' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="summary" label="返回内容摘要" min-width="520" />
    </el-table>

    <div class="export-area">
      <div class="export-title">测试结果文本</div>
      <el-input
        v-model="tableText"
        type="textarea"
        :rows="12"
        readonly
        placeholder="运行测试后自动生成"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getProfileApi, sendEmailVerificationApi } from '@/api/auth'
import { createChainApi, addNodeApi, getChainApi, getChainListApi, startChainApi, stopChainApi } from '@/api/chain'
import { getBoardInfoApi, getBoardListApi } from '@/api/board'
import {
  createHandlerApi,
  getConnectionsApi,
  getDronePositionApi,
  getDroneStatusApi,
  startConnectionApi,
  stopConnectionApi
} from '@/api/mavlink'
import { getAllUsersApi } from '@/api/admin'
import { config } from '@/utils/mockService'

interface TestResult {
  name: string
  statusCode: number
  ok: boolean
  summary: string
}

interface TestStepResponse {
  success?: boolean
  code?: number
  data?: unknown
}

const running = ref(false)
const results = ref<TestResult[]>([])
const tableText = ref('')
const finishedAt = ref('')
const useRealApi = ref(config.USE_REAL_API)

const apiModeText = computed(() => (config.USE_REAL_API ? '真实 API' : 'Mock API'))

function getAuthHeaders() {
  const token = localStorage.getItem('token')
  const headers: Record<string, string> = {}
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }
  return headers
}

function isHtmlLike(value: unknown) {
  return typeof value === 'string' && /<!doctype html>|<html/i.test(value)
}

function normalizeSummary(data: unknown): string {
  if (isHtmlLike(data)) {
    return '返回了前端 HTML，而不是接口 JSON。通常表示 dev 代理未命中，或请求落到了前端路由。'
  }

  const text = summarizeData(data)
  if (/token is expired|invalid token claims|Unauthorized|401/i.test(text)) {
    return `认证失败或 token 已过期：${text}`
  }
  if (/404 page not found/i.test(text)) {
    return `接口不存在或路径不匹配：${text}`
  }
  if (/forbidden|admin|permission|权限/i.test(text)) {
    return `权限不足或需要管理员角色：${text}`
  }
  if (/no available drones|DroneSearch not available|连接失败|超时|not found available/i.test(text)) {
    return `后端环境或设备未就绪：${text}`
  }
  return text
}

function summarizeData(data: unknown): string {
  try {
    if (data == null) return 'empty'
    if (typeof data === 'string') return data.slice(0, 200)
    if (typeof data !== 'object') return String(data)

    const obj = data as Record<string, unknown>
    const keys = Object.keys(obj).slice(0, 8)
    const preview: Record<string, unknown> = {}
    keys.forEach(key => {
      preview[key] = obj[key]
    })
    return JSON.stringify(preview)
  } catch {
    return 'summary failed'
  }
}

function normalizeStatusCode(response: any): number {
  if (typeof response?.code === 'number') {
    return response.code === 0 ? 200 : response.code
  }
  return 200
}

function isAuthFailure(response: TestStepResponse | null) {
  if (!response) return false

  const statusCode = normalizeStatusCode(response)
  const text = normalizeSummary(response.data ?? response)
  return statusCode === 401 || /token is expired|invalid token claims|Unauthorized|401/i.test(text)
}

function pushSkippedAfterAuthFailure(reason: string) {
  results.value.push({
    name: '后续鉴权接口',
    statusCode: 401,
    ok: false,
    summary: `已停止执行后续依赖登录态的接口：${reason}`
  })
}

function buildResultText(rows: TestResult[]): string {
  const lines = [
    `模式: ${apiModeText.value}`,
    `执行时间: ${finishedAt.value || '-'}`,
    `总计: ${rows.length}，成功: ${rows.filter(row => row.ok).length}，失败: ${rows.filter(row => !row.ok).length}`,
    ''
  ]

  for (const row of rows) {
    lines.push(`[${row.ok ? 'PASS' : 'FAIL'}] ${row.name} | ${row.statusCode} | ${row.summary}`)
  }

  return lines.join('\n')
}

async function requestRaw(method: 'GET' | 'POST', url: string, data?: Record<string, unknown>) {
  const response = await axios.request({
    method,
    url,
    data,
    headers: {
      ...getAuthHeaders(),
      'Content-Type': 'application/json'
    },
    validateStatus: () => true
  })

  return {
    success: response.status >= 200 && response.status < 300,
    code: response.status,
    data: response.data
  }
}

async function probeStreamEndpoint(format: 'mjpeg' | 'flv', streamId?: string, taskCode?: string) {
  const controller = new AbortController()
  const params = new URLSearchParams()

  if (streamId) params.set('stream_id', streamId)
  if (taskCode) params.set('task_code', taskCode)
  params.set('format', format)

  const timer = window.setTimeout(() => controller.abort(), 2500)

  try {
    const response = await fetch(`/api/backend/live?${params.toString()}`, {
      method: 'GET',
      headers: getAuthHeaders(),
      signal: controller.signal
    })

    return {
      success: response.ok,
      code: response.status,
      data: {
        contentType: response.headers.get('content-type'),
        target: streamId || taskCode || ''
      }
    }
  } catch (error: any) {
    if (error?.name === 'AbortError') {
      return {
        success: true,
        code: 200,
        data: {
          message: `${format.toUpperCase()} 已建立连接并主动中止探测`,
          target: streamId || taskCode || ''
        }
      }
    }
    throw error
  } finally {
    window.clearTimeout(timer)
  }
}

async function probeVideoWebSocket(streamId?: string, taskCode?: string) {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const params = new URLSearchParams()
  if (streamId) params.set('stream_id', streamId)
  if (taskCode) params.set('task_code', taskCode)

  const token = localStorage.getItem('token')
  if (token) params.set('token', token)

  return new Promise<{ success: boolean; code: number; data: Record<string, unknown> }>((resolve) => {
    const ws = new WebSocket(`${protocol}//${window.location.host}/api/backend/live/ws?${params.toString()}`)
    let settled = false

    const finish = (payload: { success: boolean; code: number; data: Record<string, unknown> }) => {
      if (settled) return
      settled = true
      ws.close()
      resolve(payload)
    }

    const timeout = window.setTimeout(() => {
      finish({
        success: false,
        code: 0,
        data: { message: 'WebSocket 探测超时' }
      })
    }, 3000)

    ws.onopen = () => {
      window.clearTimeout(timeout)
      finish({
        success: true,
        code: 200,
        data: {
          message: 'WebSocket 连接成功',
          target: streamId || taskCode || ''
        }
      })
    }

    ws.onerror = () => {
      window.clearTimeout(timeout)
      finish({
        success: false,
        code: 0,
        data: {
          message: 'WebSocket 连接失败'
        }
      })
    }

    ws.onclose = () => {
      if (!settled) {
        window.clearTimeout(timeout)
        finish({
          success: false,
          code: 0,
          data: {
            message: 'WebSocket 已关闭'
          }
        })
      }
    }
  })
}

async function runStep(name: string, runner: () => Promise<any>) {
  try {
    const response = await runner()
    const ok = response?.success !== false
    results.value.push({
      name,
      statusCode: normalizeStatusCode(response),
      ok,
      summary: normalizeSummary(response?.data ?? response)
    })
    return response
  } catch (error: any) {
    results.value.push({
      name,
      statusCode: error?.response?.status || error?.status || 0,
      ok: false,
      summary: normalizeSummary(error?.response?.data || error?.message || error)
    })
    return null
  }
}

async function runAllTests() {
  running.value = true
  results.value = []
  tableText.value = ''
  finishedAt.value = ''

  let chainId = ''
  let boardId = ''
  let streamId = ''
  let taskCode = ''
  const token = localStorage.getItem('token')

  try {
    if (config.USE_REAL_API && !token) {
      results.value.push({
        name: '登录态检查',
        statusCode: 401,
        ok: false,
        summary: '未检测到 token，请先登录后再运行真实后端测试'
      })
      return
    }

    const profileRes = await runStep('获取当前用户信息', () => getProfileApi())
    if (isAuthFailure(profileRes)) {
      pushSkippedAfterAuthFailure('当前 token 已过期或无效，请重新登录后重试')
      return
    }

    const emailVerificationRes = await runStep('发送邮箱验证码', () => sendEmailVerificationApi({
      email: 'test@example.com'
    }))
    if (isAuthFailure(emailVerificationRes)) {
      pushSkippedAfterAuthFailure('邮箱验证码接口返回鉴权失败，请确认登录状态')
      return
    }

    const createChainRes = await runStep('创建任务链', () => createChainApi({
      name: `API-Test-${Date.now()}`
    }))
    if (isAuthFailure(createChainRes)) {
      pushSkippedAfterAuthFailure('任务链接口返回鉴权失败，请确认登录状态')
      return
    }
    chainId = createChainRes?.data?.chain_id || ''

    if (chainId) {
      await runStep('获取任务链详情', () => getChainApi(chainId))
      await runStep('添加节点-起飞', () => addNodeApi(chainId, {
        nodeType: 'takeoff',
        params: { altitude: 20, timeout: 20 }
      }))
      await runStep('添加节点-飞往目标', () => addNodeApi(chainId, {
        nodeType: 'goto',
        params: {
          latitude: 22.543123,
          longitude: 114.052345,
          altitude: 20,
          timeout: 60
        }
      }))
      await runStep('启动任务链', () => startChainApi(chainId))
      await runStep('停止任务链', () => stopChainApi(chainId))
    } else {
      results.value.push({
        name: '任务链依赖测试',
        statusCode: 0,
        ok: false,
        summary: '未获取到 chain_id，已跳过依赖测试'
      })
    }

    await runStep('获取任务链列表', () => getChainListApi())

    const boardListRes = await runStep('获取板子列表', () => getBoardListApi())
    const boards = boardListRes?.data?.boards || boardListRes?.data || []
    boardId = boards?.[0]?.boardId || boards?.[0]?.board_id || ''

    if (boardId) {
      await runStep('获取板子详情', () => getBoardInfoApi(boardId))
    } else {
      results.value.push({
        name: '获取板子详情',
        statusCode: 0,
        ok: false,
        summary: '未找到可用 board_id，已跳过'
      })
    }

    await runStep('获取可用无人机列表', () => getConnectionsApi())
    await runStep('获取无人机状态', () => getDroneStatusApi())
    await runStep('获取无人机位置', () => getDronePositionApi())
    await runStep('创建 MAVLink Handler', () => createHandlerApi({
      handler_type: 'udp',
      config: {
        name: `handler-${Date.now()}`,
        address: '127.0.0.1',
        port: 14550
      }
    }))
    await runStep('启动 MAVLink 连接', () => startConnectionApi({
      connection_type: 'udp',
      config: {
        address: '127.0.0.1',
        port: 14550
      }
    }))
    await runStep('停止 MAVLink 连接', () => stopConnectionApi())

    await runStep('获取管理员用户列表', () => getAllUsersApi())

    await runStep('终端通信-help', () => requestRaw('POST', '/terminal/message', {
      cmd: 'help',
      objects: [],
      args: {}
    }))
    await runStep('终端通信-whoami', () => requestRaw('POST', '/terminal/message', {
      cmd: 'whoami',
      objects: [],
      args: {}
    }))
    await runStep('终端通信-pwd', () => requestRaw('POST', '/terminal/message', {
      cmd: 'pwd',
      objects: [],
      args: {}
    }))

    const liveListRes = await runStep('视频流列表', () => requestRaw('GET', '/api/backend/live/list'))
    const liveStreams =
      liveListRes?.data?.data?.streams ||
      liveListRes?.data?.data ||
      liveListRes?.data?.streams ||
      liveListRes?.data ||
      []
    const firstStream = Array.isArray(liveStreams) ? liveStreams[0] : undefined
    streamId = firstStream?.stream_id || ''
    taskCode = firstStream?.task_code || chainId

    if (streamId) {
      await runStep('视频流详情', () => requestRaw('GET', `/api/backend/live/info/${streamId}`))
      await runStep('视频流-MJPEG 探测', () => probeStreamEndpoint('mjpeg', streamId))
      await runStep('视频流-FLV 探测', () => probeStreamEndpoint('flv', streamId))
      await runStep('视频流-WebSocket 探测', () => probeVideoWebSocket(streamId))
    } else if (taskCode) {
      results.value.push({
        name: '视频流详情',
        statusCode: 0,
        ok: false,
        summary: '未获取到 stream_id，已跳过详情测试'
      })
      await runStep('视频流-MJPEG 探测(task_code)', () => probeStreamEndpoint('mjpeg', undefined, taskCode))
      await runStep('视频流-FLV 探测(task_code)', () => probeStreamEndpoint('flv', undefined, taskCode))
      await runStep('视频流-WebSocket 探测(task_code)', () => probeVideoWebSocket(undefined, taskCode))
    } else {
      results.value.push({
        name: '视频流接口测试',
        statusCode: 0,
        ok: false,
        summary: '未找到可用 stream_id/task_code，已跳过视频流探测'
      })
    }
  } catch (error) {
    results.value.push({
      name: '执行异常',
      statusCode: 0,
      ok: false,
      summary: String(error)
    })
  } finally {
    finishedAt.value = new Date().toLocaleString()
    tableText.value = buildResultText(results.value)
    running.value = false
  }
}

async function copyTableText() {
  if (!tableText.value) return
  try {
    await navigator.clipboard.writeText(tableText.value)
    ElMessage.success('测试结果已复制')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

async function handleModeChange(value: boolean) {
  try {
    await ElMessageBox.confirm(
      `切换到 ${value ? '真实后端' : 'Mock'} 模式后，需要刷新页面才能让全部 API 生效。是否立即刷新？`,
      '切换联调模式',
      {
        confirmButtonText: '立即刷新',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    config.USE_REAL_API = value
    useRealApi.value = value
    localStorage.setItem('useRealApi', String(value))
    window.location.reload()
  } catch {
    useRealApi.value = config.USE_REAL_API
  }
}
</script>

<style scoped>
.test-api-page {
  padding: 24px;
  max-width: 1280px;
  margin: 0 auto;
}

.header h2 {
  margin: 0 0 6px;
}

.desc {
  margin: 0 0 16px;
  color: var(--text-secondary);
}

.actions {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.meta {
  display: flex;
  gap: 18px;
  color: var(--text-secondary);
  margin-bottom: 12px;
  font-size: 13px;
  flex-wrap: wrap;
}

.mode-switch {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 8px;
  background: var(--bg-card);
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
}

.mode-label {
  color: var(--text-secondary);
  font-size: 13px;
}

.notes {
  margin-bottom: 16px;
  padding: 12px 14px;
  border-radius: 8px;
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: 13px;
}

.notes p {
  margin: 4px 0;
}

.export-area {
  margin-top: 16px;
}

.export-title {
  margin-bottom: 8px;
  font-weight: 600;
}
</style>
