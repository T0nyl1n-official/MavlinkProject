#include <Arduino.h>
#include <Wire.h>
#include <Adafruit_BMP280.h>
#include <Adafruit_MLX90614.h>
#include <FastLED.h>
#include <WiFi.h>
#include <time.h>

// ========== WiFi 配置 ==========
const char *ssid = "YILING_10";
const char *password = "yiling10";

// ========== TCP 配置（原 UDP 改为 TCP）==========
const char *tcpAddress = "172.67.148.20";
const int tcpPort = 8081;

// ========== 设备标识配置 ==========
const String FROM_ID = "esp32_c3_001";
const String FROM_TYPE = "Drone";
const String TO_ID = "Control";
const String TO_TYPE = "Control";

// ========== 消息配置（与后端结构体保持一致）==========
const String MESSAGE_TYPE = "Request";
const String MESSAGE_ATTRIBUTE = "Status";
const String CONNECTION = "TCP";  // 改为 TCP
const String COMMAND = "Heartbeat";

// ========== 引脚定义 ==========
#define MQ4_PIN 2
#define WS2812B_PIN 7
#define I2C_SDA 4
#define I2C_SCL 5
#define NUM_LEDS 48

// ========== 传感器对象 ==========
Adafruit_BMP280 bmp;
Adafruit_MLX90614 mlx;
CRGB leds[NUM_LEDS];

// ========== TCP 客户端对象 ==========
WiFiClient tcpClient;

// ========== 传感器数据变量 ==========
uint32_t gasValue = 0;
float temperature = 0;
float pressure = 0;
float altitude = 0;
float ambientTemp = 0;
float objectTemp = 0;

// ========== WiFi 状态变量 ==========
bool lastWiFiConnected = false;
unsigned long lastWiFiCheck = 0;
unsigned long lastBlinkTime = 0;
bool wifiBlinkState = false;
int wifiBlinkCount = 0;

// ========== TCP 连接管理变量 ==========
bool tcpConnected = false;
unsigned long lastTcpReconnectTime = 0;
const unsigned long TCP_RECONNECT_INTERVAL = 5000;  // 5秒重连一次

// ========== 初始化 LED ==========
void initLED()
{
    FastLED.addLeds<WS2812B, WS2812B_PIN, GRB>(leds, NUM_LEDS);
    FastLED.setBrightness(50);
    FastLED.clear();
    FastLED.show();
}

// ========== WiFi 状态指示 LED ==========
void updateLEDByWiFiStatus()
{
    bool wifiConnected = (WiFi.status() == WL_CONNECTED);
    
    if (wifiConnected != lastWiFiConnected) {
        if (!wifiConnected) {
            wifiBlinkCount = 6;
            Serial.println("⚠️ WiFi 断开，触发指示灯闪烁提醒");
        }
        lastWiFiConnected = wifiConnected;
    }
    
    if (wifiConnected && wifiBlinkCount == 0) {
        uint8_t red = constrain(map(gasValue, 0, 3000, 0, 255), 0, 255);
        uint8_t green = constrain(map(gasValue, 0, 3000, 255, 0), 0, 255);
        uint8_t blue = 0;
        
        if (objectTemp > 60.0 || objectTemp < -20.0) {
            blue = 128;
            Serial.println("⚠️ 温度异常，增加蓝色警示");
        }
        
        fill_solid(leds, NUM_LEDS, CRGB(red, green, blue));
        FastLED.show();
        return;
    }
    
    unsigned long now = millis();
    
    if (wifiBlinkCount > 0) {
        if (now - lastBlinkTime >= 300) {
            lastBlinkTime = now;
            wifiBlinkState = !wifiBlinkState;
            wifiBlinkCount--;
            
            if (wifiBlinkState) {
                fill_solid(leds, NUM_LEDS, CRGB::Blue);
            } else {
                fill_solid(leds, NUM_LEDS, CRGB::Black);
            }
            FastLED.show();
        }
    } else {
        if (!wifiConnected) {
            static unsigned long lastFlashTime = 0;
            static bool flashState = false;
            
            if (now - lastFlashTime >= 500) {
                lastFlashTime = now;
                flashState = !flashState;
                
                if (flashState) {
                    fill_solid(leds, NUM_LEDS, CRGB(0, 0, 150));
                } else {
                    fill_solid(leds, NUM_LEDS, CRGB::Black);
                }
                FastLED.show();
            }
        }
    }
}

void updateLED()
{
    updateLEDByWiFiStatus();
}

// ========== 读取 MQ-4 ==========
void readMQ4()
{
    gasValue = analogRead(MQ4_PIN);
}

// ========== 读取 BMP280 ==========
bool readBMP280()
{
    temperature = bmp.readTemperature();
    pressure = bmp.readPressure() / 100.0F;
    altitude = bmp.readAltitude(1013.25);
    
    if (isnan(temperature) || isnan(pressure)) {
        Serial.println("BMP280 读取失败");
        return false;
    }
    return true;
}

// ========== 读取 MLX90614 ==========
bool readMLX90614()
{
    ambientTemp = mlx.readAmbientTempC();
    objectTemp = mlx.readObjectTempC();
    
    if (isnan(ambientTemp) || isnan(objectTemp)) {
        Serial.println("MLX90614 读取失败");
        return false;
    }
    return true;
}

// ========== 连接 WiFi ==========
void connectWiFi()
{
    Serial.println("正在连接到 WiFi...");
    WiFi.mode(WIFI_STA);
    WiFi.begin(ssid, password);
    
    int attempts = 0;
    while (WiFi.status() != WL_CONNECTED && attempts < 20) {
        delay(500);
        Serial.print(".");
        attempts++;
    }
    
    if (WiFi.status() == WL_CONNECTED) {
        Serial.println();
        Serial.println("WiFi 已连接");
        Serial.print("IP 地址: ");
        Serial.println(WiFi.localIP());
    } else {
        Serial.println();
        Serial.println("WiFi 连接失败");
    }
}

// ========== WiFi 掉线重连 ==========
void ensureWiFiConnected()
{
    if (WiFi.status() == WL_CONNECTED)
        return;
    
    Serial.println("WiFi 连接已断开，正在重新连接...");
    connectWiFi();
}

// ========== TCP 连接管理 ==========
void connectTCP()
{
    if (tcpClient.connected()) {
        return;
    }
    
    Serial.println("正在连接 TCP 服务器...");
    Serial.print("目标地址: ");
    Serial.print(tcpAddress);
    Serial.print(":");
    Serial.println(tcpPort);
    
    if (tcpClient.connect(tcpAddress, tcpPort)) {
        tcpConnected = true;
        Serial.println("✓ TCP 服务器连接成功");
    } else {
        tcpConnected = false;
        Serial.println("✗ TCP 服务器连接失败");
    }
}

void ensureTCPConnected()
{
    // 检查 TCP 连接状态
    if (!tcpClient.connected()) {
        tcpConnected = false;
        
        // 限制重连频率
        unsigned long now = millis();
        if (now - lastTcpReconnectTime >= TCP_RECONNECT_INTERVAL) {
            lastTcpReconnectTime = now;
            connectTCP();
        }
    } else {
        tcpConnected = true;
    }
}

// ========== 初始化 NTP ==========
void initNTP()
{
    configTime(8 * 3600, 0, "ntp.aliyun.com", "pool.ntp.org", "ntp.tencent.com");
}

// ========== 生成唯一 MessageID ==========
String generateMessageID()
{
    return String("esp32_") + String(millis()) + "_" + String(random(1000, 9999));
}

// ========== 获取 RFC3339 时间 ==========
String getCurrentTime()
{
    struct tm timeinfo;
    if (!getLocalTime(&timeinfo))
    {
        return "2026-01-01T00:00:00Z";
    }
    
    char buffer[30];
    strftime(buffer, sizeof(buffer), "%Y-%m-%dT%H:%M:%SZ", &timeinfo);
    return String(buffer);
}

// ========== 构建 Data 对象 ==========
String buildDataJson()
{
    String dataJson = "{";
    dataJson += "\"status\":\"idle\",";
    dataJson += "\"gas\":" + String(gasValue) + ",";
    dataJson += "\"temperature\":" + String(temperature, 2) + ",";
    dataJson += "\"pressure\":" + String(pressure, 2) + ",";
    dataJson += "\"altitude\":" + String(altitude, 2) + ",";
    dataJson += "\"ambient_temp\":" + String(ambientTemp, 2) + ",";
    dataJson += "\"object_temp\":" + String(objectTemp, 2) + ",";
    dataJson += "\"wifi_rssi\":" + String(WiFi.RSSI());
    dataJson += "}";
    return dataJson;
}

// ========== 构建内层 Message 对象 ==========
String buildInnerMessage()
{
    String innerMsg = "{";
    innerMsg += "\"message_type\":\"" + MESSAGE_TYPE + "\",";
    innerMsg += "\"message_attribute\":\"" + MESSAGE_ATTRIBUTE + "\",";
    innerMsg += "\"connection\":\"" + CONNECTION + "\",";
    innerMsg += "\"command\":\"" + COMMAND + "\",";
    innerMsg += "\"data\":" + buildDataJson();
    innerMsg += "}";
    return innerMsg;
}

// ========== 构建外层 BoardMessage 对象 ==========
String buildMessageJson()
{
    String messageId = generateMessageID();
    String messageTime = getCurrentTime();
    
    String messageJson = "{";
    messageJson += "\"message_id\":\"" + messageId + "\",";
    messageJson += "\"message_time\":\"" + messageTime + "\",";
    messageJson += "\"message\":" + buildInnerMessage() + ",";
    messageJson += "\"from_id\":\"" + FROM_ID + "\",";
    messageJson += "\"from_type\":\"" + FROM_TYPE + "\",";
    messageJson += "\"to_id\":\"" + TO_ID + "\",";
    messageJson += "\"to_type\":\"" + TO_TYPE + "\"";
    messageJson += "}";
    return messageJson;
}

// ========== 通过 TCP 发送 BoardMessage ==========
void sendDataViaTCP()
{
    if (WiFi.status() != WL_CONNECTED) {
        Serial.println("WiFi 未连接，无法发送数据");
        return;
    }
    
    if (!tcpClient.connected()) {
        Serial.println("TCP 未连接，无法发送数据");
        return;
    }
    
    String messageJson = buildMessageJson();
    
    Serial.println("========== 准备发送 TCP 消息 ==========");
    Serial.print("目标地址: ");
    Serial.print(tcpAddress);
    Serial.print(":");
    Serial.println(tcpPort);
    Serial.print("消息内容: ");
    Serial.println(messageJson);
    Serial.println("=====================================");
    
    // 发送数据（添加换行符作为消息分隔符）
    tcpClient.println(messageJson);
    
    Serial.println("✓ TCP 消息已成功发送");
}

// ========== 打印调试 ==========
void printDebug()
{
    Serial.print("GAS:");
    Serial.print(gasValue);
    Serial.print(" | TEMP:");
    Serial.print(temperature);
    Serial.print("C | PRESS:");
    Serial.print(pressure);
    Serial.print("hPa | ALT:");
    Serial.print(altitude);
    Serial.print("m | AMB:");
    Serial.print(ambientTemp);
    Serial.print("C | OBJ:");
    Serial.print(objectTemp);
    Serial.print("C | WiFi:");
    Serial.print(WiFi.status() == WL_CONNECTED ? "OK" : "NO");
    Serial.print(" | TCP:");
    Serial.print(tcpConnected ? "OK" : "NO");
    Serial.print(" | RSSI:");
    Serial.println(WiFi.RSSI());
}

// ========== setup ==========
void setup()
{
    Serial.begin(115200);
    delay(500);
    
    Serial.println();
    Serial.println("========== ESP32 启动 ==========");
    
    Wire.begin(I2C_SDA, I2C_SCL);
    
    initLED();
    Serial.println("LED 已初始化");
    
    if (!bmp.begin(0x76)) {
        Serial.println("BMP280 初始化失败！尝试地址 0x77");
        if (!bmp.begin(0x77)) {
            Serial.println("BMP280 初始化失败，请检查接线");
        } else {
            Serial.println("BMP280 已初始化（地址 0x77）");
        }
    } else {
        Serial.println("BMP280 已初始化（地址 0x76）");
    }
    
    mlx.begin();
    Serial.println("MLX90614 已初始化");
    
    initNTP();
    Serial.println("NTP 已初始化");
    
    connectWiFi();
    
    // 启动提示 LED 闪烁（3次黄色）
    for (int i = 0; i < 3; i++) {
        fill_solid(leds, NUM_LEDS, CRGB::Yellow);
        FastLED.show();
        delay(300);
        fill_solid(leds, NUM_LEDS, CRGB::Black);
        FastLED.show();
        delay(300);
    }
    
    Serial.println("========== 初始化完成 ==========");
}

// ========== loop ==========
void loop()
{
    ensureWiFiConnected();
    
    // 确保 TCP 连接
    if (WiFi.status() == WL_CONNECTED) {
        ensureTCPConnected();
    }
    
    unsigned long now = millis();
    if (now - lastWiFiCheck >= 500) {
        lastWiFiCheck = now;
        bool currentWiFiState = (WiFi.status() == WL_CONNECTED);
        if (currentWiFiState != lastWiFiConnected) {
            updateLED();
        }
    }
    
    readMQ4();
    readBMP280();
    readMLX90614();
    updateLED();
    printDebug();
    
    // 通过 TCP 发送数据（仅当连接正常时）
    if (WiFi.status() == WL_CONNECTED && tcpClient.connected()) {
        sendDataViaTCP();
    } else {
        Serial.println("⚠️ TCP 未连接，跳过数据发送");
    }
    
    delay(1000);
}