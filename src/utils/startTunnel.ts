import { spawn, type ChildProcess } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, unlinkSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

const TUNNEL_NAME = 'my-frontend'
const TARGET_URL = 'http://localhost:3002'
const DEFAULT_PUBLIC_URL = 'https://frontend-api.deeppluse.dpdns.org'

let tunnelProcess: ChildProcess | null = null

function log(message: string) {
  console.log(`[Tunnel] ${message}`)
}

function logError(message: string) {
  console.error(`[Tunnel] ${message}`)
}

function getCacheDir() {
  const cacheDir = join(process.cwd(), '.deploy-tunnel')
  if (!existsSync(cacheDir)) {
    mkdirSync(cacheDir, { recursive: true })
  }
  return cacheDir
}

function getPidFile() {
  return join(getCacheDir(), 'frontend-tunnel.pid')
}

function readStoredPid() {
  if (!existsSync(getPidFile())) return null

  try {
    const raw = readFileSync(getPidFile(), 'utf8').trim()
    const pid = Number(raw)
    return Number.isFinite(pid) && pid > 0 ? pid : null
  } catch {
    return null
  }
}

function writePidFile(pid: number) {
  writeFileSync(getPidFile(), String(pid), 'utf8')
}

function removePidFile() {
  try {
    if (existsSync(getPidFile())) {
      unlinkSync(getPidFile())
    }
  } catch {
    // ignore
  }
}

function isProcessRunning(pid: number) {
  try {
    process.kill(pid, 0)
    return true
  } catch {
    return false
  }
}

function extractPublicUrl(output: string) {
  const matchedUrl = output.match(/https:\/\/[^\s]+/)
  return matchedUrl?.[0] ?? null
}

function handleTunnelOutput(chunk: Buffer | string) {
  const text = chunk.toString().trim()
  if (!text) return

  const publicUrl = extractPublicUrl(text)
  if (publicUrl) {
    log(`公网地址: ${publicUrl}`)
  }

  if (/address already in use|bind: address already in use/i.test(text)) {
    logError('端口被占用，请检查本地服务或已有 tunnel 进程')
    return
  }

  if (/failed to connect to origin|connection refused|dial tcp .*3002/i.test(text)) {
    logError('本地 3002 端口不可用，请先启动前端服务')
    return
  }

  console.log(text)
}

function attachProcessEvents(child: ChildProcess) {
  child.stdout?.on('data', chunk => {
    handleTunnelOutput(chunk)
  })

  child.stderr?.on('data', chunk => {
    handleTunnelOutput(chunk)
  })

  child.on('error', error => {
    const errorCode = (error as NodeJS.ErrnoException).code
    if (errorCode === 'ENOENT') {
      logError('未检测到 cloudflared，请先安装并加入 PATH')
      logError('安装说明：https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/')
      return
    }

    logError(`启动失败：${error.message}`)
  })

  child.on('exit', () => {
    tunnelProcess = null
    removePidFile()
    log('隧道已停止')
  })
}

export function startTunnel() {
  if (tunnelProcess?.pid && isProcessRunning(tunnelProcess.pid)) {
    log(`隧道已在运行，PID=${tunnelProcess.pid}`)
    return tunnelProcess
  }

  const existingPid = readStoredPid()
  if (existingPid && isProcessRunning(existingPid)) {
    log(`检测到已有隧道进程，PID=${existingPid}`)
    return null
  }

  const child = spawn(
    'cloudflared',
    ['tunnel', '--url', TARGET_URL, 'run', TUNNEL_NAME],
    {
      cwd: process.cwd(),
      stdio: ['ignore', 'pipe', 'pipe'],
      windowsHide: true,
      detached: false
    }
  )

  tunnelProcess = child
  attachProcessEvents(child)

  if (child.pid) {
    writePidFile(child.pid)
    log('启动成功')
    log(`公网地址: ${process.env.VITE_TUNNEL_PUBLIC_URL || DEFAULT_PUBLIC_URL}`)
  }

  return child
}

export function stopTunnel() {
  if (tunnelProcess?.pid && isProcessRunning(tunnelProcess.pid)) {
    tunnelProcess.kill('SIGTERM')
    log('已发送停止指令')
    return
  }

  const existingPid = readStoredPid()
  if (existingPid && isProcessRunning(existingPid)) {
    process.kill(existingPid, 'SIGTERM')
    removePidFile()
    log('已停止已运行的隧道进程')
  }
}
