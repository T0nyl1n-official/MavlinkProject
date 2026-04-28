# ESP32-C3 天然气检测报警系统

## 硬件清单
- ESP32-C3-DevKitM-1
- MQ-4 天然气传感器
- GY-906 红外测温传感器
- BMP280 气压传感器
- WS2812B 灯带（48颗）
- MB102 电源模块 + 12V适配器

## 接线说明

| 传感器 | 引脚 | 接ESP32 |
|--------|------|---------|
| MQ-4 | AO | GPIO2 |
| GY-906 | SDA/SCL | GPIO4/GPIO5 |
| BMP280 | SDA/SCL | GPIO4/GPIO5 |
| 灯带 | DIN | GPIO7 |

**供电**：MB102 5V→灯带+MQ-4，3.3V→GY-906+BMP280，所有GND共地

## API接口

```
POST https://api.deeppluse.dpdns.org/api/sensor/message
```

```json
{
  "sensor_id": "esp32_001",
  "alert_type": "gas_leak",
  "latitude": 31.23,
  "longitude": 121.47,
  "timestamp": 1713177600,
  "severity": "high"
}
```

## 灯带颜色

| 颜色 | 含义 |
|------|------|
| 绿色 | 安全 |
| 橙色 | 气体警告 |
| 红色 | 严重泄漏 |
| 蓝色 | 温度异常 |

## 运行

1. 修改 WiFi 名称和密码
2. 编译上传
3. 预热 MQ-4 约5-10分钟
