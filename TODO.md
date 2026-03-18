#测试前设置环境变量
$env:MavlinkMysqlDSN="root:123456@tcp(127.0.0.1:3306)/mavlink_db?charset=utf8mb4&parseTime=True&loc=Local"

# 接口测试待办清单 (API Todo List)

> **前置条件**: 所有请求 Header 必须包含 `Authorization: Bearer <Token>`

## 1. 任务链模块 (Chain Module)
- Base URL: `/api/chain`

- [x] **创建任务链 (Create Chain)**
    - Method: `POST` /create
    - Body: `{"name": "Test Chain 01"}`
    - 预期: 200 OK, 返回 `chain_id`

- [x] **获取任务链列表 (List Chains)**
    - Method: `GET` /list
    - 预期: 200 OK, 返回刚才创建的链

- [ ] **获取任务链详情 (Get Chain)**
    - Method: `GET` /:id
    - 预期: 200 OK

- [x] **添加节点 (Add Node)**
    - Method: `POST` /:id/node/add
    - Body:
      ```json
      {
        "node_type": "CheckPoint", 
        "params": {
           "wait_time": 5
        }
      }
      ```
      *(注: node_type 可能的值: createHandler, connectionStart, droneTakeoff 等)*
    - 预期: 200 OK, 返回 `node_id`

- [ ] **删除节点 (Delete Node)**
    - Method: `POST` /:id/node/delete/:nodeId
    - 预期: 200 OK

- [x] **执行控制 (Control)**
    - [x] `POST` /:id/start
    - [ ] `POST` /:id/pause
    - [ ] `POST` /:id/stop
    - [ ] `POST` /:id/reset
    - 预期: 200 OK, observe chain status change

- [x] **删除任务链 (Delete Chain)**
    - Method: `DELETE` /:id
    - 预期: 200 OK

## 2. Mavlink V1 基础控制 (Basic Control)
- Base URL: `/mavlink/v1`

- [x] **创建连接句柄 (Create Handler)** - **Critical**
    - Method: `POST` /handler/create
    - Body:
      ```json
      {
        "connection_type": "udp",
        "udp_addr": "127.0.0.1",
        "udp_port": 14550,
        "system_id": 1,
        "component_id": 1,
        "protocol_version": "2.0"
      }
      ```
    - 预期: 200 OK, **务必记下返回的 `handler_id`**

- [x] **连接控制**
    - [x] `POST` /connection/start?handler_id=<YOUR_ID>
    - [ ] `POST` /connection/stop?handler_id=<YOUR_ID>
    - [x] `GET` /handler/:id (查看状态)

- [x] **无人机动作 (Action)**
    *(需配合 SITL 模拟器，否则可能超时或无响应)*
    - [x] `POST` /drone/takeoff?handler_id=<YOUR_ID> & Body: `{"altitude": 10}`
    - [ ] `POST` /drone/land
    - [ ] `POST` /drone/mode (Body: `{"mode": "GUIDED"}`)
    - [ ] `POST` /drone/return

- [ ] **无人机状态 (Telemetry)**
    - [ ] `GET` /drone/status
    - [ ] `GET` /drone/position
    - [ ] `GET` /drone/battery

## 3. Mavlink V2 高级控制 (Advanced)
- Base URL: `/mavlink/v2`
- **注意**: 依赖 V1 中创建的 `handler_id`

- [x] **起飞 (Takeoff V2)**
    - Method: `POST` /takeoff
    - Query: `?handler_id=<YOUR_ID>`
    - Body: `{"altitude": 20.5}`

- [ ] **降落 (Land V2)**
    - Method: `POST` /land
    - Body: `{"latitude": 0, "longitude": 0}` (如果需要指定地点)

- [x] **移动 (Move)**
    - Method: `POST` /move
    - Body: `{"x": 10, "y": 10, "z": -5}`

- [ ] **返航充电 (Return To Charge)** - *Unique Feature*
    - Method: `POST` /return-charge
    - 预期: 检查是否触发自动充电逻辑