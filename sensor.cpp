#include <Arduino.h>
#include <Wire.h>
#include <Adafruit_BMP280.h>
#include <Adafruit_MLX90614.h>
#include <FastLED.h>
#include <WiFi.h>
#include <HTTPClient.h>
#include <WiFiClientSecure.h>
#include <time.h>

// ========== WiFi 配置 ==========
const char *ssid = "";
const char *password = "";

// ========== API 配置 ==========
const char *apiEndpoint = "https://api.deeppluse.dpdns.org/api/sensor/message";

// ========== 设备配置 ==========
const String SENSOR_ID = "esp32_001";
const String SENSOR_NAME = "ESP32-C3_Gas_Sensor";

// ========== 固定经纬度（设备安装位置）==========
const float LATITUDE = 31.2304;      // 纬度
const float LONGITUDE = 121.4737;    // 经度

// ========== 报警阈值配置 ==========
const int GAS_WARNING_THRESHOLD = 500;     // 气体浓度警告阈值
const int GAS_ALERT_THRESHOLD = 1000;      // 气体浓度报警阈值
const float TEMP_HIGH_THRESHOLD = 60.0;    // 高温阈值（°C）
const float TEMP_LOW_THRESHOLD = -20.0;    // 低温阈值（°C）

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

// ========== 报警类型变量 ==========
String alertType = "none";
String alertMsg = "";
String severity = "low";

// ========== WiFi 状态变量 ==========
bool lastWiFiConnected = false;
unsigned long lastWiFiCheck = 0;
unsigned long lastBlinkTime = 0;
bool wifiBlinkState = false;
int wifiBlinkCount = 0;

// ========== 发送间隔变量 ==========
unsigned long lastSendTime = 0;
const unsigned long SEND_INTERVAL = 1000;  // 秒发送一次

// ========== 初始化 LED ==========
void initLED()
{
    FastLED.addLeds<WS2812B, WS2812B_PIN, GRB>(leds, NUM_LEDS);
    FastLED.setBrightness(50);
    FastLED.clear();
    FastLED.show();
}

// ========== 判断报警类型和严重程度 ==========
void checkAlertType()
{
    // 默认无报警
    alertType = "none";
    alertMsg = "";
    severity = "low";
    
    // 气体浓度报警（优先级最高）
    if (gasValue >= GAS_ALERT_THRESHOLD) {
        alertType = "gas_leak";
        alertMsg = "气体严重泄漏！浓度: " + String(gasValue);
        severity = "high";
        Serial.println("?? 气体严重报警！");
    }
    // 气体浓度警告
    else if (gasValue >= GAS_WARNING_THRESHOLD) {
        alertType = "gas_warning";
        alertMsg = "气体泄漏警告，浓度: " + String(gasValue);
        severity = "medium";
        Serial.println("?? 气体警告！");
    }
    // 温度过高报警
    else if (objectTemp >= TEMP_HIGH_THRESHOLD) {
        alertType = "overheat";
        alertMsg = "温度过高！物体温度: " + String(objectTemp) + "°C";
        severity = "high";
        Serial.println("?? 温度过高报警！");
    }
    // 温度过低报警
    else if (objectTemp <= TEMP_LOW_THRESHOLD) {
        alertType = "freeze";
        alertMsg = "温度过低！物体温度: " + String(objectTemp) + "°C";
        severity = "medium";
        Serial.println("?? 温度过低报警！");
    }
}

// ========== WiFi 状态指示 LED ==========
void updateLEDByWiFiStatus()
{
    bool wifiConnected = (WiFi.status() == WL_CONNECTED);
    
    if (wifiConnected != lastWiFiConnected) {
        if (!wifiConnected) {
            wifiBlinkCount = 6;
            Serial.println("?? WiFi 断开，触发指示灯闪烁提醒");
        }
        lastWiFiConnected = wifiConnected;
    }
    
    if (wifiConnected && wifiBlinkCount == 0) {
        // 根据报警类型显示不同颜色
        if (alertType == "gas_leak") {
            fill_solid(leds, NUM_LEDS, CRGB::Red);
        } else if (alertType == "gas_warning") {
            fill_solid(leds, NUM_LEDS, CRGB(255, 165, 0));
        } else if (alertType == "overheat") {
            fill_solid(leds, NUM_LEDS, CRGB::Blue);
        } else {
            uint8_t red = constrain(map(gasValue, 0, 3000, 0, 255), 0, 255);
            uint8_t green = constrain(map(gasValue, 0, 3000, 255, 0), 0, 255);
            fill_solid(leds, NUM_LEDS, CRGB(red, green, 0));
        }
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

// ========== 获取当前时间戳（秒）==========
uint32_t getTimestamp()
{
    struct tm timeinfo;
    if (!getLocalTime(&timeinfo)) {
        return millis() / 1000;
    }
    return mktime(&timeinfo);
}

// ========== 构建 JSON（按后端 API 格式）==========
String buildRequestJson()
{
    String json = "{";
    json += "\"sensor_id\":\"" + SENSOR_ID + "\",";
    json += "\"sensor_ip\":\"" + WiFi.localIP().toString() + "\",";
    json += "\"sensor_name\":\"" + SENSOR_NAME + "\",";
    json += "\"alert_type\":\"" + alertType + "\",";
    json += "\"alert_msg\":\"" + alertMsg + "\",";
    json += "\"latitude\":" + String(LATITUDE, 6) + ",";
    json += "\"longitude\":" + String(LONGITUDE, 6) + ",";
    json += "\"timestamp\":" + String(getTimestamp()) + ",";
    json += "\"severity\":\"" + severity + "\"";
    json += "}";
    return json;
}

// ========== HTTPS POST 请求 ==========
void sendDataToBackend()
{
    if (WiFi.status() != WL_CONNECTED) {
        Serial.println("WiFi 未连接，无法发送数据");
        return;
    }
    
    String jsonData = buildRequestJson();
    
    Serial.println("========== 发送数据到后端 ==========");
    Serial.print("POST: ");
    Serial.println(apiEndpoint);
    Serial.print("消息内容: ");
    Serial.println(jsonData);
    
    WiFiClientSecure client;
    client.setInsecure();
    
    HTTPClient http;
    http.begin(client, apiEndpoint);
    http.addHeader("Content-Type", "application/json");
    
    int httpCode = http.POST(jsonData);
    
    if (httpCode > 0) {
        Serial.print("? HTTPS POST 成功，状态码: ");
        Serial.println(httpCode);
        String response = http.getString();
        Serial.print("服务器响应: ");
        Serial.println(response);
    } else {
        Serial.print("? HTTPS POST 失败，错误码: ");
        Serial.println(httpCode);
    }
    
    http.end();
    Serial.println("====================================");
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
    Serial.print(" | ALERT:");
    Serial.print(alertType);
    Serial.print(" | SEV:");
    Serial.println(severity);
}

// ========== setup ==========
void setup()
{
    Serial.begin(115200);
    delay(500);
    
    Serial.println();
    Serial.println("╔══════════════════════════════════════╗");
    Serial.println("║     ESP32-C3 传感器报警系统启动      ║");
    Serial.println("╚══════════════════════════════════════╝");
    Serial.println();
    
    Wire.begin(I2C_SDA, I2C_SCL);
    
    initLED();
    Serial.println("? LED 已初始化");
    
    if (!bmp.begin(0x76)) {
        Serial.println("? BMP280 初始化失败！尝试地址 0x77");
        if (!bmp.begin(0x77)) {
            Serial.println("? BMP280 初始化失败，请检查接线");
        } else {
            Serial.println("? BMP280 已初始化（地址 0x77）");
        }
    } else {
        Serial.println("? BMP280 已初始化（地址 0x76）");
    }
    
    mlx.begin();
    Serial.println("? MLX90614 已初始化");
    
    initNTP();
    Serial.println("? NTP 已初始化");
    
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
    Serial.println();
}

// ========== loop ==========
void loop()
{
    ensureWiFiConnected();
    
    // 读取传感器数据
    readMQ4();
    readBMP280();
    readMLX90614();
    
    // 判断报警类型
    checkAlertType();
    
    // 更新 LED 颜色
    updateLED();
    
    // 打印调试信息
    printDebug();
    
    // 定期发送数据到后端
    unsigned long now = millis();
    if (now - lastSendTime >= SEND_INTERVAL) {
        lastSendTime = now;
        if (WiFi.status() == WL_CONNECTED) {
            sendDataToBackend();
        } else {
            Serial.println("?? WiFi 未连接，跳过数据发送");
        }
    }
    
    delay(100);
}
