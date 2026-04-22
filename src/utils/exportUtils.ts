/**
 * 数据导出工具
 * 支持 JSON 和 CSV 格式的数据导出
 */

/**
 * 导出数据为 JSON 文件
 * @param data 要导出的数据
 * @param filename 文件名
 */
export function exportToJSON(data: any, filename: string) {
    try {
        const jsonStr = JSON.stringify(data, null, 2)
        const blob = new Blob([jsonStr], { type: 'application/json' })
        const url = URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = `${filename}.json`
        link.click()
        URL.revokeObjectURL(url)
    } catch (error) {
        console.error('导出 JSON 失败:', error)
        throw new Error('导出 JSON 文件失败')
    }
}

/**
 * 导出数据为 CSV 文件
 * @param data 要导出的数据（数组）
 * @param filename 文件名
 * @param headers 自定义表头
 */
export function exportToCSV(data: any[], filename: string, headers?: Record<string, string>) {
    try {
        if (!data || data.length === 0) {
            throw new Error('没有数据可导出')
        }

        // 生成表头
        const defaultHeaders = Object.keys(data[0])
        const csvHeaders = headers ? Object.values(headers) : defaultHeaders
        const csvHeaderRow = csvHeaders.join(',')

        // 生成数据行
        const csvRows = data.map(item => {
            return defaultHeaders
                .map(header => {
                    const value = item[header]
                    // 处理包含逗号、引号或换行符的值
                    if (typeof value === 'string' && (value.includes(',') || value.includes('"') || value.includes('\n'))) {
                        return `"${value.replace(/"/g, '""')}"`
                    }
                    return value
                })
                .join(',')
        })

        // 组合 CSV 内容
        const csvContent = [csvHeaderRow, ...csvRows].join('\n')
        const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
        const url = URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = `${filename}.csv`
        link.click()
        URL.revokeObjectURL(url)
    } catch (error) {
        console.error('导出 CSV 失败:', error)
        throw new Error('导出 CSV 文件失败')
    }
}

/**
 * 导出任务链数据
 * @param chains 任务链数据
 * @param format 导出格式
 */
export function exportChains(chains: any[], format: 'json' | 'csv') {
    const filename = `任务链数据_${new Date().toISOString().split('T')[0]}`

    if (format === 'json') {
        exportToJSON(chains, filename)
    } else {
        const headers = {
            id: 'ID',
            name: '名称',
            status: '状态',
            nodes: '节点数',
            created_at: '创建时间',
            updated_at: '更新时间'
        }

        // 转换数据格式
        const formattedChains = chains.map(chain => ({
            id: chain.id,
            name: chain.name,
            status: chain.status,
            nodes: chain.nodes?.length || 0,
            created_at: chain.created_at,
            updated_at: chain.updated_at
        }))

        exportToCSV(formattedChains, filename, headers)
    }
}

/**
 * 导出板子数据
 * @param boards 板子数据
 * @param format 导出格式
 */
export function exportBoards(boards: any[], format: 'json' | 'csv') {
    const filename = `板子数据_${new Date().toISOString().split('T')[0]}`

    if (format === 'json') {
        exportToJSON(boards, filename)
    } else {
        const headers = {
            id: 'ID',
            name: '名称',
            ip: 'IP地址',
            port: '端口',
            is_connected: '连接状态',
            last_heartbeat: '最后心跳'
        }

        // 转换数据格式
        const formattedBoards = boards.map(board => ({
            id: board.id,
            name: board.name,
            ip: board.ip,
            port: board.port,
            is_connected: board.is_connected ? '在线' : '离线',
            last_heartbeat: board.last_heartbeat
        }))

        exportToCSV(formattedBoards, filename, headers)
    }
}

/**
 * 导出错误日志
 * @param logs 错误日志数据
 * @param format 导出格式
 */
export function exportErrorLogs(logs: any[], format: 'json' | 'csv') {
    const filename = `错误日志_${new Date().toISOString().split('T')[0]}`

    if (format === 'json') {
        exportToJSON(logs, filename)
    } else {
        const headers = {
            id: 'ID',
            time: '时间',
            type: '类型',
            message: '消息',
            source: '来源'
        }

        exportToCSV(logs, filename, headers)
    }
}
