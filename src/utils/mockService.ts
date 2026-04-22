// Mock 服务，集中管理所有前端模拟数据
export const config = {
    USE_REAL_API: false
}

export interface MockUser {
    User_ID: number
    Username: string
    Email: string
    Role: string
    token: string
}

export interface MockChain {
    chain_id: string
    chain_name: string
    description: string
    nodes: MockNode[]
    status: 'idle' | 'running' | 'completed' | 'failed'
    created_at: string
}

export interface MockNode {
    node_id: string
    node_type: string
    node_name: string
    parameters: Record<string, any>
    position: number
}

export interface MockBoard {
    board_id: string
    board_name: string
    board_type: string
    connection: string
    address: string
    port: string
    is_connected: boolean
}

class MockService {
    private users: MockUser[] = []
    private chains: MockChain[] = []
    private boards: MockBoard[] = []
    private terminalHistory: { cmd: string; out: string }[] = []

    constructor() {
        // 初始化默认数据
        this.initializeData()
    }

    private initializeData() {
        // 初始化 Mock 板子数据
        this.boards = [
            {
                board_id: 'board-001',
                board_name: '飞控板 1',
                board_type: 'Pixhawk',
                connection: 'UDP',
                address: '192.168.1.100',
                port: '14550',
                is_connected: true
            },
            {
                board_id: 'board-002',
                board_name: '飞控板 2',
                board_type: 'Ardupilot',
                connection: 'TCP',
                address: '192.168.1.101',
                port: '5760',
                is_connected: false
            }
        ]

        // 初始化 Mock 任务链数据
        this.chains = [
            {
                chain_id: 'chain-001',
                chain_name: '起飞任务',
                description: '测试起飞任务链',
                nodes: [
                    {
                        node_id: 'node-001',
                        node_type: 'takeoff',
                        node_name: '起飞',
                        parameters: { altitude: 10 },
                        position: 0
                    },
                    {
                        node_id: 'node-002',
                        node_type: 'waypoint',
                        node_name: '航点 1',
                        parameters: { latitude: 31.2304, longitude: 121.4737, altitude: 10 },
                        position: 1
                    }
                ],
                status: 'idle',
                created_at: new Date().toISOString()
            }
        ]
    }

    // 登录 Mock
    login(email: string, _password: string): Promise<MockUser> {
        return new Promise((resolve) => {
            setTimeout(() => {
                const token = `mock-token-${Date.now()}`
                const user: MockUser = {
                    User_ID: Date.now(),
                    Username: email.split('@')[0],
                    Email: email,
                    Role: 'admin',
                    token
                }
                this.users.push(user)
                resolve(user)
            }, 500)
        })
    }

    // 注册 Mock
    register(username: string, email: string, _password: string): Promise<MockUser> {
        return new Promise((resolve) => {
            setTimeout(() => {
                const token = `mock-token-${Date.now()}`
                const user: MockUser = {
                    User_ID: Date.now(),
                    Username: username,
                    Email: email,
                    Role: 'user',
                    token
                }
                this.users.push(user)
                resolve(user)
            }, 500)
        })
    }

    // 获取用户信息 Mock
    getProfile(): Promise<MockUser> {
        return new Promise((resolve) => {
            setTimeout(() => {
                const user = this.users[0] || {
                    User_ID: 1,
                    Username: 'mock-user',
                    Email: 'mock@example.com',
                    Role: 'admin',
                    token: 'mock-token'
                }
                resolve(user)
            }, 300)
        })
    }

    // 任务链相关 Mock
    getChains(): Promise<MockChain[]> {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve(this.chains)
            }, 300)
        })
    }

    createChain(name: string, description: string): Promise<MockChain> {
        return new Promise((resolve) => {
            setTimeout(() => {
                const chain: MockChain = {
                    chain_id: `chain-${Date.now()}`,
                    chain_name: name,
                    description,
                    nodes: [],
                    status: 'idle',
                    created_at: new Date().toISOString()
                }
                this.chains.push(chain)
                resolve(chain)
            }, 500)
        })
    }

    addNode(chainId: string, nodeType: string, nodeName: string, parameters: Record<string, any>): Promise<MockNode> {
        return new Promise((resolve) => {
            setTimeout(() => {
                const chain = this.chains.find(c => c.chain_id === chainId)
                if (chain) {
                    const node: MockNode = {
                        node_id: `node-${Date.now()}`,
                        node_type: nodeType,
                        node_name: nodeName,
                        parameters,
                        position: chain.nodes.length
                    }
                    chain.nodes.push(node)
                    resolve(node)
                } else {
                    throw new Error('Chain not found')
                }
            }, 300)
        })
    }

    startChain(chainId: string): Promise<{ success: boolean }> {
        return new Promise((resolve) => {
            setTimeout(() => {
                const chain = this.chains.find(c => c.chain_id === chainId)
                if (chain) {
                    chain.status = 'running'
                }
                resolve({ success: true })
            }, 500)
        })
    }

    stopChain(chainId: string): Promise<{ success: boolean }> {
        return new Promise((resolve) => {
            setTimeout(() => {
                const chain = this.chains.find(c => c.chain_id === chainId)
                if (chain) {
                    chain.status = 'idle'
                }
                resolve({ success: true })
            }, 300)
        })
    }

    deleteNode(chainId: string, nodeId: string): Promise<{ success: boolean }> {
        return new Promise((resolve) => {
            setTimeout(() => {
                const chain = this.chains.find(c => c.chain_id === chainId)
                if (chain) {
                    const index = chain.nodes.findIndex(n => n.node_id === nodeId)
                    if (index > -1) {
                        chain.nodes.splice(index, 1)
                        // 更新剩余节点的位置
                        chain.nodes.forEach((node, i) => {
                            node.position = i
                        })
                    }
                }
                resolve({ success: true })
            }, 300)
        })
    }

    // 板子相关 Mock
    getBoards(): Promise<MockBoard[]> {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve(this.boards)
            }, 300)
        })
    }

    createBoard(board: Omit<MockBoard, 'board_id' | 'is_connected'>): Promise<MockBoard> {
        return new Promise((resolve) => {
            setTimeout(() => {
                const newBoard: MockBoard = {
                    ...board,
                    board_id: `board-${Date.now()}`,
                    is_connected: false
                }
                this.boards.push(newBoard)
                resolve(newBoard)
            }, 500)
        })
    }

    deleteBoard(boardId: string): Promise<{ success: boolean }> {
        return new Promise((resolve) => {
            setTimeout(() => {
                const index = this.boards.findIndex(b => b.board_id === boardId)
                if (index > -1) {
                    this.boards.splice(index, 1)
                }
                resolve({ success: true })
            }, 300)
        })
    }

    sendBoardCommand(boardId: string, command: string): Promise<{ success: boolean; message: string }> {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve({
                    success: true,
                    message: `Command "${command}" sent to board ${boardId}`
                })
            }, 300)
        })
    }

    // 终端相关 Mock
    executeTerminalCommand(cmd: string): Promise<string> {
        return new Promise((resolve) => {
            setTimeout(() => {
                let response = ''

                if (cmd.includes('get status')) {
                    response = 'System status: OK\nUptime: 10 hours\nCPU usage: 15%\nMemory usage: 45%'
                } else if (cmd.includes('get settings')) {
                    response = 'Current settings:\n- UDP Port: 14550\n- Log Level: info\n- Max Connections: 100'
                } else if (cmd.includes('list chains')) {
                    response = 'Available chains:\n1. chain-001: 起飞任务 (idle)'
                } else {
                    response = `[Mock] Command received: ${cmd}\nThis is a mock response. In production, this would execute on the backend.`
                }

                this.terminalHistory.push({ cmd, out: response })
                resolve(response)
            }, 300)
        })
    }

    // MAVLink 相关 Mock
    getMavlinkDevices(): Promise<Array<{ id: string; version: string; ip: string; port: number; sysid: number; compid: number; connected: boolean }>> {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve([
                    {
                        id: 'mavlink-001',
                        version: 'v1.0',
                        ip: '192.168.1.100',
                        port: 14550,
                        sysid: 1,
                        compid: 1,
                        connected: true
                    }
                ])
            }, 300)
        })
    }

    connectMavlink(_deviceId: string): Promise<{ success: boolean }> {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve({ success: true })
            }, 500)
        })
    }

    disconnectMavlink(_deviceId: string): Promise<{ success: boolean }> {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve({ success: true })
            }, 300)
        })
    }
}

// 导出单例
export const mockService = new MockService()


