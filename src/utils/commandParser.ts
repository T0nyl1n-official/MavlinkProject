/**
 * 命令解析器
 * 用于解析用户输入的命令行指令
 */

/**
 * 解析后的命令结构
 */
export interface ParsedCommand {
  command: string
  params: Record<string, string | number | boolean>
  raw: string
}

/**
 * 解析命令行输入
 * @param input 用户输入的命令字符串
 * @returns 解析后的命令对象
 */
export function parseCommand(input: string): ParsedCommand | null {
  const trimmed = input.trim()
  if (!trimmed) return null

  const parts = trimmed.split(/\s+/)
  const command = parts[0].toLowerCase()
  const params: Record<string, string | number | boolean> = {}

  // 解析 --xxx 格式的参数
  for (let i = 1; i < parts.length; i++) {
    const part = parts[i]
    if (part.startsWith('--')) {
      const paramName = part.slice(2)
      const nextPart = parts[i + 1]
      if (nextPart && !nextPart.startsWith('--')) {
        params[paramName] = isNaN(Number(nextPart)) ? nextPart : Number(nextPart)
        i++
      } else {
        params[paramName] = true
      }
    }
  }

  return { command, params, raw: input }
}

/**
 * 支持的命令列表
 */
const SUPPORTED_COMMANDS = ['takeoff', 'land', 'move', 'shutdown', 'status', 'help']

/**
 * 验证命令是否有效
 * @param command 命令名称
 * @returns 是否为有效命令
 */
export function isValidCommand(command: string): boolean {
  return SUPPORTED_COMMANDS.includes(command)
}

/**
 * 获取命令帮助信息
 * @param command 命令名称（可选）
 * @returns 帮助信息
 */
export function getCommandHelp(command?: string): string {
  if (command && !isValidCommand(command)) {
    return `未知命令: ${command}`
  }

  const helpText: Record<string, string> = {
    takeoff: '起飞无人机\n用法: takeoff [--alt 高度]\n参数:\n  --alt  起飞高度（米），默认为10米',
    land: '降落无人机\n用法: land',
    move: '移动无人机\n用法: move [--lat 纬度] [--lng 经度] [--alt 高度]\n参数:\n  --lat  目标纬度\n  --lng  目标经度\n  --alt  目标高度（米）',
    shutdown: '关闭无人机\n用法: shutdown',
    status: '查看无人机状态\n用法: status',
    help: '显示帮助信息\n用法: help [命令]'
  }

  if (command) {
    return helpText[command] || `没有找到命令 "${command}" 的帮助信息`
  }

  return `可用命令:\n${SUPPORTED_COMMANDS.map(cmd => `  ${cmd}`).join('\n')}\n\n使用 "help <命令>" 查看具体命令的帮助信息`
}
