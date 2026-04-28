/**
 * Cloudflare Tunnel 一键部署工具（仅 Node.js 环境使用）
 *
 * 用法：
 * - 在 vite.config.ts 中：`import { maybeAutoStartDeployTunnel } from './src/utils/deployTunnel.ts'`
 * - 手动：`import { startDeployTunnel, stopDeployTunnel } from '@/utils/deployTunnel'`
 *
 * 环境变量：
 * - DEPLOY_TUNNEL_AUTO：设为 `1` / `true` 时在 maybeAutoStartDeployTunnel 中自动启动
 * - DEPLOY_TUNNEL_LOCAL_PORT：本地服务端口，默认 3002
 * - DEPLOY_TUNNEL_HOSTNAME：对外域名，默认 frontend-api.deeppluse.dpdns.org
 * - CLOUDFLARED_TUNNEL_ID：命名隧道 UUID（与 credentials 配套，用于自定义 hostname）
 * - CLOUDFLARED_CREDENTIALS_FILE：credentials.json 绝对或相对路径
 * - CLOUDFLARED_BIN：cloudflared 可执行文件名，默认 cloudflared
 *
 * 未配置命名隧道时，将使用快速隧道：`cloudflared tunnel --url http://127.0.0.1:<port>`
 * （此时不会绑定 frontend-api.deeppluse.dpdns.org，控制台会给出说明）
 */

import { spawn, type ChildProcess } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, unlinkSync, writeFileSync } from 'node:fs'
import { createConnection } from 'node:net'
import { join } from 'node:path'
import { pathToFileURL } from 'node:url'

export type DeployEnvironment = 'development' | 'production'

export interface DeployTunnelOptions {
  /** 本地前端（或预览）监听端口，默认读取 DEPLOY_TUNNEL_LOCAL_PORT 或 3002 */
  localPort?: number
  /** 对外 hostname，默认 frontend-api.deeppluse.dpdns.org */
  hostname?: string
  /** cloudflared 可执行文件，默认 cloudflared */
  cloudflaredBin?: string
}

export type DeployTunnelStartResult =
  | { ok: true; pid: number; mode: 'named-config' | 'quick' | 'skipped'; message: string }
  | { ok: false; reason: string }

const DEFAULT_LOCAL_PORT = Number(process.env.DEPLOY_TUNNEL_LOCAL_PORT ?? '3002') || 3002
const DEFAULT_HOSTNAME =
  process.env.DEPLOY_TUNNEL_HOSTNAME ?? 'frontend-api.deeppluse.dpdns.org'

let activeChild: ChildProcess | null = null

function getCacheDir(): string {
  const dir = join(process.cwd(), '.deploy-tunnel')
  if (!existsSync(dir)) {
    mkdirSync(dir, { recursive: true })
  }
  return dir
}

function getPidPath(): string {
  return join(getCacheDir(), 'cloudflared.pid')
}

function getConfigPath(): string {
  return join(getCacheDir(), 'cloudflared-config.yml')
}

/** 判断当前是否为 Node.js 运行时（排除被打进浏览器的情况） */
function isNodeRuntime(): boolean {
  return typeof process !== 'undefined' && Boolean(process.versions?.node)
}

/** 检测当前环境：开发 / 生产（基于 NODE_ENV、MODE） */
export function detectDeployEnvironment(): DeployEnvironment {
  if (!isNodeRuntime()) return 'development'
  if (process.env.NODE_ENV === 'production') return 'production'
  const mode = process.env.MODE ?? process.env.VITE_MODE
  if (mode === 'production') return 'production'
  return 'development'
}

function readPidFromFile(): number | null {
  const pidPath = getPidPath()
  if (!existsSync(pidPath)) return null
  try {
    const raw = readFileSync(pidPath, 'utf8').trim()
    const n = Number(raw)
    return Number.isFinite(n) && n > 0 ? n : null
  } catch {
    return null
  }
}

function writePidFile(pid: number): void {
  writeFileSync(getPidPath(), String(pid), 'utf8')
}

function removePidFile(): void {
  try {
    if (existsSync(getPidPath())) unlinkSync(getPidPath())
  } catch {
    /* ignore */
  }
}

/** 进程是否仍存在（POSIX / Windows 均可用 signal 0 探测） */
export function isProcessAlive(pid: number): boolean {
  try {
    process.kill(pid, 0)
    return true
  } catch {
    return false
  }
}

function log(prefix: string, message: string): void {
  // eslint-disable-next-line no-console
  console.log(`[deployTunnel] ${prefix} ${message}`)
}

function logError(prefix: string, message: string): void {
  // eslint-disable-next-line no-console
  console.error(`[deployTunnel] ${prefix} ${message}`)
}

/** 检查 cloudflared 是否已由本工具启动且仍在运行 */
export function isDeployTunnelRunning(): boolean {
  if (activeChild && !activeChild.killed && activeChild.pid) {
    return isProcessAlive(activeChild.pid)
  }
  const pid = readPidFromFile()
  if (pid == null) return false
  return isProcessAlive(pid)
}

function resolveCredentialsPathForYaml(credentialsPath: string): string {
  const trimmed = credentialsPath.trim()
  const isAbsolute =
    trimmed.startsWith('/') || /^[A-Za-z]:[/\\]/.test(trimmed)
  const resolved = isAbsolute ? trimmed : join(process.cwd(), trimmed)
  return resolved.replace(/\\/g, '/')
}

function credentialsFileExists(credentialsPath: string): boolean {
  const trimmed = credentialsPath.trim()
  const isAbsolute =
    trimmed.startsWith('/') || /^[A-Za-z]:[/\\]/.test(trimmed)
  const abs = isAbsolute ? trimmed : join(process.cwd(), trimmed)
  return existsSync(abs)
}

function buildNamedTunnelConfig(opts: {
  tunnelId: string
  credentialsFile: string
  hostname: string
  localPort: number
}): string {
  const credPath = resolveCredentialsPathForYaml(opts.credentialsFile)
  return [
    `tunnel: ${opts.tunnelId}`,
    `credentials-file: ${credPath}`,
    'ingress:',
    `  - hostname: ${opts.hostname}`,
    `    service: http://127.0.0.1:${opts.localPort}`,
    '  - service: http_status:404',
    ''
  ].join('\n')
}

/** 检测本地端口是否已有进程监听（用于提示端口占用或前端未启动） */
export function probeLocalPort(port: number): Promise<{ listening: boolean; error?: string }> {
  return new Promise(resolve => {
    const socket = createConnection({ host: '127.0.0.1', port }, () => {
      socket.destroy()
      resolve({ listening: true })
    })
    socket.setTimeout(800)
    socket.on('error', err => {
      const code = (err as NodeJS.ErrnoException).code
      if (code === 'ECONNREFUSED') {
        resolve({ listening: false })
      } else if (code === 'EADDRINUSE') {
        resolve({ listening: false, error: `端口 ${port} 地址占用（EADDRINUSE）` })
      } else {
        resolve({ listening: false, error: err.message })
      }
    })
    socket.on('timeout', () => {
      socket.destroy()
      resolve({ listening: false, error: '探测本地端口超时' })
    })
  })
}

function attachChildLogging(child: ChildProcess): void {
  child.stdout?.on('data', chunk => {
    log('[cloudflared]', chunk.toString().trimEnd())
  })
  child.stderr?.on('data', chunk => {
    const text = chunk.toString()
    const lower = text.toLowerCase()
    if (lower.includes('address already in use') || lower.includes('bind: address already in use')) {
      logError('失败', `端口冲突：${text.trim()}`)
    } else {
      logError('[cloudflared]', text.trimEnd())
    }
  })
  child.on('exit', (code, signal) => {
    log('退出', `cloudflared 已结束 code=${code} signal=${signal ?? 'none'}`)
    removePidFile()
    if (activeChild === child) activeChild = null
  })
  child.on('error', err => {
    logError('失败', `子进程错误: ${err.message}`)
    if ((err as NodeJS.ErrnoException).code === 'ENOENT') {
      logError(
        '提示',
        '未找到 cloudflared，请先安装并加入 PATH：https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/'
      )
    }
  })
}

/**
 * 启动 cloudflared tunnel，将流量转发到本地端口（默认 3002）。
 * 不阻塞调用方：内部使用 spawn，立即返回。
 */
export async function startDeployTunnel(
  options: DeployTunnelOptions = {}
): Promise<DeployTunnelStartResult> {
  if (!isNodeRuntime()) {
    return { ok: false, reason: 'deployTunnel 仅在 Node.js 环境可用' }
  }

  const env = detectDeployEnvironment()
  const localPort = options.localPort ?? DEFAULT_LOCAL_PORT
  const hostname = options.hostname ?? DEFAULT_HOSTNAME
  const bin = options.cloudflaredBin ?? process.env.CLOUDFLARED_BIN ?? 'cloudflared'

  log('环境', `${env}（NODE_ENV=${process.env.NODE_ENV ?? 'unset'}, MODE=${process.env.MODE ?? 'unset'}）`)

  const existingPid = readPidFromFile()
  if (existingPid != null && isProcessAlive(existingPid)) {
    log('跳过', `检测到 tunnel 已在运行 PID=${existingPid}，不重复启动`)
    return { ok: true, pid: existingPid, mode: 'skipped', message: '已在运行（PID 文件）' }
  }
  if (activeChild?.pid && isProcessAlive(activeChild.pid)) {
    log('跳过', `当前进程已持有 cloudflared 子进程 PID=${activeChild.pid}，不重复启动`)
    return { ok: true, pid: activeChild.pid, mode: 'skipped', message: '已在运行（当前子进程）' }
  }

  const portProbe = await probeLocalPort(localPort)
  if (!portProbe.listening) {
    log(
      '警告',
      `127.0.0.1:${localPort} 当前无服务监听，tunnel 启动后访问可能失败。${portProbe.error ?? ''}`
    )
  }

  const tunnelId = process.env.CLOUDFLARED_TUNNEL_ID?.trim()
  const credentialsFile = process.env.CLOUDFLARED_CREDENTIALS_FILE?.trim()

  let args: string[]
  let mode: 'named-config' | 'quick'

  if (tunnelId && credentialsFile) {
    if (!credentialsFileExists(credentialsFile)) {
      return {
        ok: false,
        reason: `找不到 credentials 文件：${credentialsFile}（请检查 CLOUDFLARED_CREDENTIALS_FILE）`
      }
    }
    const yaml = buildNamedTunnelConfig({
      tunnelId,
      credentialsFile,
      hostname,
      localPort
    })
    const configPath = getConfigPath()
    writeFileSync(configPath, yaml, 'utf8')
    args = ['tunnel', '--config', configPath, 'run']
    mode = 'named-config'
    log('配置', `已写入 ${pathToFileURL(configPath).href}`)
    log('路由', `${hostname} -> http://127.0.0.1:${localPort}`)
  } else {
    args = ['tunnel', '--url', `http://127.0.0.1:${localPort}`]
    mode = 'quick'
    log(
      '提示',
      '未设置 CLOUDFLARED_TUNNEL_ID + CLOUDFLARED_CREDENTIALS_FILE，使用快速隧道（随机 trycloudflare 域名）。若需绑定 frontend-api.deeppluse.dpdns.org，请在 Cloudflare 创建命名隧道并配置上述环境变量。'
    )
  }

  try {
    const child = spawn(bin, args, {
      detached: false,
      stdio: ['ignore', 'pipe', 'pipe'],
      env: { ...process.env },
      windowsHide: true
    })
    activeChild = child
    attachChildLogging(child)

    if (!child.pid) {
      return { ok: false, reason: 'spawn 成功但未分配 PID' }
    }
    writePidFile(child.pid)
    log('启动', `已启动 cloudflared PID=${child.pid} mode=${mode}`)
    return {
      ok: true,
      pid: child.pid,
      mode,
      message:
        mode === 'named-config'
          ? `tunnel 已启动，对外域名（ingress）: ${hostname}`
          : '快速隧道已启动，请查看上方 cloudflared 输出的 trycloudflare URL'
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    logError('失败', `启动 cloudflared 异常: ${msg}`)
    return { ok: false, reason: msg }
  }
}

/** 停止由本工具启动的 cloudflared（优先结束当前子进程，其次按 PID 文件 kill） */
export async function stopDeployTunnel(): Promise<void> {
  if (!isNodeRuntime()) return

  if (activeChild && !activeChild.killed && activeChild.pid) {
    try {
      activeChild.kill('SIGTERM')
      log('停止', `已发送 SIGTERM 至 PID=${activeChild.pid}`)
    } catch (e) {
      logError('停止', e instanceof Error ? e.message : String(e))
    }
    activeChild = null
  }

  const pid = readPidFromFile()
  if (pid != null && isProcessAlive(pid)) {
    try {
      process.kill(pid, 'SIGTERM')
      log('停止', `已向 PID 文件中的进程 ${pid} 发送 SIGTERM`)
    } catch (e) {
      logError('停止', e instanceof Error ? e.message : String(e))
    }
  }
  removePidFile()
}

/**
 * 根据环境变量自动启动（供 vite.config 或 CI 调用）。
 * DEPLOY_TUNNEL_AUTO=1 / true 时执行 startDeployTunnel。
 */
export async function maybeAutoStartDeployTunnel(extra?: {
  /** Vite 传入的 mode，会写入 process.env.MODE 之前可用于判断 */
  mode?: string
}): Promise<void> {
  if (!isNodeRuntime()) return

  const flag = process.env.DEPLOY_TUNNEL_AUTO?.toLowerCase()
  const enabled = flag === '1' || flag === 'true' || flag === 'yes'
  if (!enabled) {
    return
  }

  if (extra?.mode) {
    process.env.MODE = extra.mode
  }

  log('自动', 'DEPLOY_TUNNEL_AUTO 已启用，正在启动 tunnel…')
  const result = await startDeployTunnel()
  if (!result.ok) {
    logError('自动', `启动失败: ${result.reason}`)
  } else {
    log('自动', result.message)
  }
}
