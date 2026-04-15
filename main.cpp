#include <Arduino.h>
#include <Wire.h>
#include <Adafruit_BMP280.h>
#include <Adafruit_MLX90614.h>
#include <FastLED.h>
#include <WiFi.h>
#include <HTTPClient.h>         // 新增：用于 HTTP 请求
#include <WiFiClientSecure.h>   // 新增：用于处理 HTTPS 传输
#include <time.h>

// ========== WiFi 配置 ==========
const char *ssid = "YILING_10";
const char *password = "yiling10";

// ========== HTTP(S) API 配置 ==========
// 这里的域名填入你做好了内网穿透或线上的真实域名
const char *apiEndpoint = "https://conn.deeppluse.dpdns.org/api/sensor/message"; 

// ========== 设备标识配置 ==========
const String FROM_ID = "esp32_001";
const String FROM_TYPE = "ESP32";

// ========== 消息配置 ==========
const String MESSAGE_TYPE = "Heartbeat";
const String MESSAGE_ATTRIBUTE = "Status";
const String CONNECTION = "HTTP";  
const String COMMAND = "Status";

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

// ========== 初始化 NTP ==========
void initNTP()
{
    configTime(8 * 3600, 0, "ntp.aliyun.com", "pool.ntp.org", "ntp.tencent.com");
}

// ========== 生成唯一 MessageID ==========
String generateMessageID()
{
    return String("msg_") + String(millis()) + "_" + String(random(1000, 9999));
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
// 将 ESP32 读出的真实传感数据组装成要求的格式
String buildDataJson()
{
    String dataJson = "{";
    dataJson += "\"temperature\":" + String(temperature, 2) + ",";
    
    // 如果你没有装湿度传感器，这里依然使用 BMP 的气压或其它你想要传的数据
    // 此处根据你的需求额外带上了真实数据供后端使用
    dataJson += "\"pressure\":" + String(pressure, 2) + ","; 
    dataJson += "\"gas\":" + String(gasValue) + ",";
    dataJson += "\"altitude\":" + String(altitude, 2) + ",";
    dataJson += "\"ambient_temp\":" + String(ambientTemp, 2) + ",";
    dataJson += "\"object_temp\":" + String(objectTemp, 2);
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
    
    String messageJson = "{";
    messageJson += "\"message_id\":\"" + messageId + "\",";
    messageJson += "\"message\":" + buildInnerMessage() + ",";
    messageJson += "\"from_id\":\"" + FROM_ID + "\",";
    messageJson += "\"from_type\":\"" + FROM_TYPE + "\"";
    messageJson += "}";
    
    return messageJson;
}

// ========== 通过 HTTPS 发送 JSON  payload ==========
void sendDataViaHTTP()
{
    if (WiFi.status() != WL_CONNECTED) {
        Serial.println("WiFi 未连接，无法发送数据");
        return;
    }
    
    String messageJson = buildMessageJson();
    
    Serial.println("========== 准备发送 HTTPS 消息 ==========");
    Serial.print("目标地址: ");
    Serial.println(apiEndpoint);
    Serial.println("消息内容: ");
    Serial.println(messageJson);
    
    // 创建针对 HTTPS 的安全客户端
    WiFiClientSecure *client = new WiFiClientSecure;
    // 不验证 TLS 证书指纹 (因为走 CF 或者各种动态代理，直接 setInsecure 是最稳定的)
    client->setInsecure(); 
    
    HTTPClient https;
    
    // 初始化 HTTPClient
    if (https.begin(*client, apiEndpoint)) {
        // 设置请求头为 JSON
        https.addHeader("Content-Type", "application/json");
        
        // 发起 POST 请求
        int httpResponseCode = https.POST(messageJson);

        if (httpResponseCode > 0) {
            Serial.print("✓ HTTPS POST 请求成功，状态码: ");
            Serial.println(httpResponseCode);
            String response = https.getString();
            Serial.println("服务器响应:");
            Serial.println(response);
        } else {
            Serial.print("✗ HTTPS 请求失败，错误代码: ");
            Serial.println(httpResponseCode);
            Serial.printf("详细错误: %s\n", https.errorToString(httpResponseCode).c_str());
        }
        
        // 关闭请求释放资源
        https.end();
    } else {
        Serial.println("✗ 无法连接到目标服务器开始 HTTPS 传输");
    }
    
    // 释放 WiFiClientSecure 对象
    delete client;
    Serial.println("=========================================");
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
    
    // 只有在 WiFi 连上的情况下发送数据
    if (WiFi.status() == WL_CONNECTED) {
        sendDataViaHTTP();
    } else {
        Serial.println("⚠️ WiFi 未连接，跳过数据发送");
    }
    
    // HTTP/HTTPS 请求对处理器资源消耗较大，建议调高延时（示例为 5 秒）
    delay(5000); 
}