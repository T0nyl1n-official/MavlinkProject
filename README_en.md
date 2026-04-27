# MavlinkProject - Drone Scheduling Management System

## Project Overview

MavlinkProject is a drone communication and scheduling management system based on the MAVLink protocol. It provides comprehensive drone control, monitoring, and scheduling capabilities, supporting multi-drone management, sensor response, progress chain execution, and real-time video streaming.

## Technology Stack

- **Backend Framework**: Gin (Go Web Framework)
- **Database**: MySQL (Data Persistence)
- **Cache**: Redis (Token Storage, Verification Codes, Error Logs)
- **Configuration**: YAML (Centralized Configuration Management) + Environment Variable Override
- **Communication**: TCP/UDP, HTTPS, FRP (Remote Port Forwarding)

## Core Function Modules

### 1. User Authentication System
- User Registration & Login
- JWT Token Authentication
- Role-based Access Control (Regular User/Admin)

### 2. Device Authentication System
- Independent Login for Hardware Devices (Central/LandNode/Sensor/Drone)
- Dedicated JWT Token Authentication
- Device Status Management (Online/Offline)
- Friendly Error Messages (No Ban Triggering)

### 3. BoardConn Communication Module
- **TCP/UDP Server**: Receives messages from Board/Sensor devices
- **MessageDispatcher**: Strategy pattern for message routing
- **Multiple Handler Support**

### 4. Handler Modules
#### Sensor Folder (Sensor-related Processing)
- `SensorAlertHandler`: Processes sensor alerts, automatically generates task chains
- `AIAgentHandler`: AI Agent extension point (reserved interface)
- `SensorMessageHandler.go`: Task chain generation logic

#### Boards Folder (Board Communication Management)
- `BoardConnection.go`: Board connection management
- `BoardHandler.go`: Flight control message processing (preserved, not in use)
- `MessageDispatcher.go`: Message dispatcher
- `MessageSender.go`: FRP message sending
- `LiveStreamHandler.go`: Real-time video stream processing

### 5. FRP Multi-Central Support
- Support for configuring multiple Central servers
- Automatic retry mechanism (configurable retry count)
- Failover: Only reports error after trying all Centrals
- **HTTPS Support**: Backend can send messages to Central via HTTPS

### 6. Progress Chain System
- Chained task execution
- Dynamic task generation
- Status tracking

### 7. Live Stream Module
- **Board/Live Interface**: Central uploads video stream
- **Backend/Live Interface**: Frontend retrieves real-time video
- **Multi-protocol Support**: MJPEG, WebSocket, RAW, FLV
- **Task Chain Association**: Video stream bound to task chain

### 8. Centralized Configuration Management
- All configuration stored in `Setting.yaml`
- Supports runtime configuration updates
- **Environment Variable Override**: Environment variables take priority, YAML used as fallback
- Modular management (Board, FRP, Server, etc.)

## Project Structure

```
MavlinkProject/
├── Server/Backend/
│   ├── Backend.go                 # Backend server entry point
│   ├── BackendAccessor.go        # Backend accessor (decoupling)
│   ├── Config/                    # Configuration management
│   │   └── SettingManager.go      # Configuration loading & memory management
│   ├── Database/                  # Database configuration
│   ├── Handler/                   # Business processors
│   │   ├── Boards/               # Board communication processor
│   │   │   ├── BoardConnection.go    # Board connection management
│   │   │   ├── BoardHandler.go        # Flight control message processing (preserved)
│   │   │   ├── MessageDispatcher.go  # Message dispatcher
│   │   │   ├── MessageSender.go      # FRP message sending
│   │   │   ├── LiveStreamHandler.go   # Live stream processing
│   │   │   └── SensorBoard/          # Sensor processing
│   │   │       ├── SensorAlertHandler.go # Sensor alert processing
│   │   │       ├── SensorMessageHandler.go # Task chain generation
│   │   │       └── types.go           # Type definitions
│   │   ├── Device/               # Device authentication handler
│   │   ├── Mavlink/              # MAVLink processor
│   │   ├── Sensor/               # Sensor handler
│   │   ├── Users/                # User handler
│   │   └── ProgressChain/        # Progress chain processor
│   ├── Middles/                   # Middleware
│   ├── Routes/                    # Route definitions
│   │   ├── Boards/              # Board routes
│   │   │   ├── BoardMessageRoute.go
│   │   │   └── LiveStreamRoutes.go   # Live stream routes
│   │   ├── Device/              # Device routes
│   │   ├── Sensor/              # Sensor routes
│   │   ├── Terminal/            # Terminal routes
│   │   └── User/                # User routes
│   ├── Shared/                    # Shared structures
│   │   ├── Boards/             # Board message format definitions
│   │   │   ├── Board_MessageFormat.go
│   │   │   └── LiveStream_Types.go    # Live stream types
│   │   ├── Device/              # Device data models
│   │   └── FRPHelper/           # FRP communication wrapper
│   │       ├── FRPHelper.go           # FRP TCP communication
│   │       └── CentralHTTPClient.go   # Central HTTPS client
│   └── Utils/                     # Utility functions
├── config/
│   └── Setting.yaml              # Centralized configuration file
├── tests/
│   └── OutputHistory/             # Test output history
├── docs/
│   ├── API.md                    # API documentation
│   ├── requirements.md           # Requirements documentation
│   ├── tech-doc.md              # Technical documentation
│   └── Frontend_LiveStream_Guide.md  # Frontend live stream guide
└── README.md
```

## System Flow

### Board Message Processing Flow
```
Sensor/Board Device
      │
      ▼ (TCP/UDP)
BoardConn Backend Listener (0.0.0.0:8081 TCP / 0.0.0.0:8082 UDP)
      │
      ▼
isSensorMessage() Check
      │
      ├─── YES ──→ MessageDispatcher ──→ SensorAlertHandler
      │                                    │
      │                                    ▼
      │                          GenerateChainAndSendToCentral
      │                                    │
      │                                    ▼ (HTTPS)
      │                              Central Server
      │                           (central.deeppluse.dpdns.org)
      │
      └─── NO ──→ BoardConn Internal Processing (messageChan)
                      │
                      ▼
               BoardHandler (preserved, not in use)
```

### Live Stream Flow
```
Central (Drone)
    │
    │ POST /api/board/live (BoardMessage + Video Binary)
    ▼
Backend (Go/Gin)
    │
    ├── Buffer video stream
    └── Forward
        │
        ├── GET /api/backend/live (MJPEG/WebSocket)
        ▼
Frontend (Vite + Vue/React)
    │
    ▼
<video> or <canvas> Real-time Display
```

### Message Type Routing

| Message Type | FromType | Command/Attribute | Route |
|--------------|----------|-------------------|--------|
| Sensor Alert | sensor/esp32/alarm | Warning/SensorAlert/Alert | → Dispatcher → SensorAlertHandler |
| Flight Control | board/drone/fc | Heartbeat/Status/Control | → BoardConn Internal |
| Video Stream | central | VideoStream | → LiveStreamHandler |

### Sensor Alert Processing Flow
```
1. Sensor detects anomaly
2. Send SensorAlert to BoardConn via TCP/UDP
3. isSensorMessage() identifies as sensor message
4. MessageDispatcher routes to SensorAlertHandler
5. SensorAlertHandler processes:
   - Parse location information
   - Generate task chain (TakeOff → GoTo → TakePhoto → Land)
   - Send to Central via CentralHTTPClient
6. Central executes task chain
```

### User Authentication Flow
```
1. POST /users/register → Register user
2. POST /users/login → Login returns JWT Token
3. Header adds Authorization: Bearer <token>
4. Access protected endpoints
```

### Device Authentication Flow
```
1. POST /device/login → Device login gets Token
2. Header adds X-Device-ID and X-Device-Type
3. Access device-exclusive endpoints
```

## Environment Variable Configuration

System prioritizes environment variables, uses YAML config or defaults when env vars are empty.

### MySQL Database Configuration
| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `MavlinkProject_backend_database_mysql_host` | MySQL Host | localhost |
| `MavlinkProject_backend_database_mysql_port` | MySQL Port | 3306 |
| `MavlinkProject_backend_database_mysql_user` | Username | root |
| `MavlinkProject_backend_database_mysql_password` | Password | (empty) |
| `MavlinkProject_backend_database_mysql_database` | Database Name | mavlinkproject |
| `MavlinkProject_backend_database_mysql_charset` | Charset | utf8mb4 |

### Redis Configuration
| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `MavlinkProject_backend_redis_host` | Redis Host | localhost |
| `MavlinkProject_backend_redis_port` | Redis Port | 6379 |
| `MavlinkProject_backend_redis_password` | Password | (empty) |

### JWT Configuration
| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `MavlinkProject_backend_jwt_secret_key` | JWT Secret | MavlinkBackendMadeByTonyl1n |

### SMTP Email Configuration
| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `SMTP_HOST` | SMTP Server | smtp.qq.com |
| `SMTP_USERNAME` | Username | (empty) |
| `SMTP_PASSWORD` | Password/Auth Code | (empty) |
| `SMTP_FROM_EMAIL` | Sender Email | (empty) |
| `SMTP_FROM_NAME` | Sender Name | MavlinkProject |

## Redis Database Allocation

| DB | Purpose |
|----|---------|
| 0 | General Warnings |
| 1 | Backend Errors |
| 2 | Frontend Errors |
| 3 | Agent Errors |
| 4 | Drone Errors |
| 5 | Sensor Errors |
| 13 | Token Storage |
| 14 | Verification Codes |

## API Endpoints Overview

### User Endpoints
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /users/register | User Registration | None |
| POST | /users/login | User Login | None |
| POST | /users/logout | User Logout | JWT |
| PUT | /users/profile | Update User Info | JWT |
| PUT | /users/password | Change Password | JWT |

### Device Endpoints
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /device/login | Device Login | None |
| POST | /device/logout | Device Logout | DeviceJWT |
| GET | /device/status | Device Status | DeviceJWT |

### Sensor Endpoints
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/sensor/message | Receive Sensor Alert | None |
| GET | /api/sensor/status | Sensor Status | None |

### Live Stream Endpoints
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/board/live | Central Upload Stream | DeviceJWT |
| GET | /api/backend/live | Frontend Get Stream | JWT |
| GET | /api/backend/live/ws | WebSocket Stream | JWT |
| GET | /api/backend/live/list | Active Stream List | JWT |

### Terminal Endpoints
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /terminal/message | Terminal Command | JWT |

## Deployment Requirements

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

## Security Features

- **JWT Token Authentication**
- **Token Storage in Redis (Supports Immediate Invalidation on Logout)**
- **Password Encryption (MD5)**
- **Role-based Access Control**
- **CORS Middleware**
- **Rate Limiting**
- **Independent Device Authentication System**
- **HTTPS Encrypted Communication**

## License

This project is for learning and research purposes only.
