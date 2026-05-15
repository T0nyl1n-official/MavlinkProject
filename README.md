# YOLOv8 热源异常检测接口文档

## 1. 服务说明

本服务将训练好的 YOLOv8s 热源异常检测模型封装为内网 HTTP API，Java 后端通过上传图片文件调用接口，算法服务返回结构化 JSON 检测结果。

- 服务框架：FastAPI
- 默认地址：`http://<算法服务IP>:8100`
- 图片传入方式：`multipart/form-data`
- 返回格式：`application/json`
- 推理预处理：服务端会先将输入图片转为灰度图，再转回三通道 BGR 图后送入 YOLO 检测
- 标注图返回：服务端会在原始图片上绘制检测框和标注，并在 `image.url` 返回标注图完整访问地址
- 默认模型：`runs/detect/thermal_detector_v2/weights/best.pt`
- 默认检测参数：`conf=0.10`，`iou=0.60`，`imgsz=1024`

## 2. 接口列表

| 接口 | 方法 | 作用 |
| --- | --- | --- |
| `/health` | `GET` | 健康检查，确认模型是否加载成功、运行设备信息 |
| `/api/v1/detect` | `POST` | 上传一张图片，返回 YOLO 检测框、置信度、灰度均值、异常等级和标注图 URL |

## 3. 健康检查接口

### 3.1 基本信息

- URL：`/health`
- Method：`GET`
- Content-Type：无
- 用途：Java 后端可在服务启动后或定时探活时调用，确认算法服务是否可用。

### 3.2 请求示例

```bash
curl http://127.0.0.1:8100/health
```

### 3.3 成功响应示例

```json
{
  "model_loaded": true,
  "model_path": "F:\\CV\\runs\\detect\\thermal_detector_v2\\weights\\best.pt",
  "device": {
    "type": "cuda",
    "cuda_available": true,
    "cuda_device_name": "NVIDIA GeForce RTX ..."
  }
}
```

### 3.4 字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `model_loaded` | boolean | 模型是否已加载 |
| `model_path` | string/null | 当前加载的模型权重路径 |
| `device.type` | string | 推理设备，可能为 `cuda` 或 `cpu` |
| `device.cuda_available` | boolean | 当前环境是否可用 CUDA |
| `device.cuda_device_name` | string/null | CUDA 显卡名称，CPU 模式下为 `null` |

## 4. 图片检测接口

### 4.1 基本信息

- URL：`/api/v1/detect`
- Method：`POST`
- Content-Type：`multipart/form-data`
- 用途：Java 后端上传一张图片，算法服务返回该图片中的热源异常检测结果。

### 4.2 请求参数

#### Query 参数

| 参数 | 类型 | 是否必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `conf` | float | 否 | `0.10` | 置信度阈值，范围 `0.0` 到 `1.0`，值越高返回结果越少 |
| `iou` | float | 否 | `0.60` | NMS IoU 阈值，范围 `0.0` 到 `1.0` |
| `imgsz` | int | 否 | `1024` | YOLO 推理图片尺寸，范围 `32` 到 `4096` |

#### Form 参数

| 参数 | 类型 | 是否必填 | 说明 |
| --- | --- | --- | --- |
| `image` | file | 是 | 待检测图片文件，支持常见图片格式，如 JPG、PNG、BMP、TIFF |

### 4.3 请求示例

```bash
curl.exe -X POST "http://127.0.0.1:8100/api/v1/detect?conf=0.10&iou=0.60" ^
  -F "image=@test_examples/00013.png"
```

### 4.4 Java 调用要点

Java 后端请使用 `multipart/form-data` 上传文件，字段名必须是 `image`。

Spring `RestTemplate` 调用示例：

```java
String url = "https://你的内网穿透地址/api/v1/detect?conf=0.10&iou=0.60";

HttpHeaders headers = new HttpHeaders();
headers.setContentType(MediaType.MULTIPART_FORM_DATA);

MultiValueMap<String, Object> body = new LinkedMultiValueMap<>();
body.add("image", new FileSystemResource("D:/test/00013.png"));

HttpEntity<MultiValueMap<String, Object>> requestEntity =
        new HttpEntity<>(body, headers);

ResponseEntity<String> response = restTemplate.postForEntity(
        url,
        requestEntity,
        String.class
);

String json = response.getBody();
```

### 4.5 成功响应示例

```json
{
  "success": true,
  "image": {
    "width": 640,
    "height": 512,
    "url": "https://你的内网穿透地址/static/annotated/2f8b5c2f6c0a4b97b7a9f927b7e2c0a1.jpg"
  },
  "detections": [
    {
      "box": {
        "xyxy": [164.33, 259.13, 207.07, 279.85],
        "xywh": [185.7, 269.49, 42.74, 20.72]
      },
      "confidence": 0.182988,
      "temperature": {
        "mean_gray": 158.8,
        "level": "HIGH Lv2"
      }
    }
  ],
  "elapsed_ms": 86.35
}
```

### 4.6 无检测结果响应示例

接口正常执行但没有检测到异常目标时，`detections` 返回空数组。灰度等级为 `NORMAL` 的检测结果不会返回，也不会在标注图中画框。

```json
{
  "success": true,
  "image": {
    "width": 640,
    "height": 512,
    "url": "https://你的内网穿透地址/static/annotated/8d80a2eea8a2468a9db2d9b07cb4d6d7.jpg"
  },
  "detections": [],
  "elapsed_ms": 52.18
}
```

### 4.7 响应字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `success` | boolean | 请求是否成功执行 |
| `image.width` | int | 原图宽度，单位像素 | -> 不确定后端需不需要，但是先写上了
| `image.height` | int | 原图高度，单位像素 |
| `image.url` | string | 标注后图片的完整访问地址，域名前缀来自 `api/server.py` 顶部的 `TUNNEL_BASE_URL` |
| `detections` | array | 异常检测结果数组，不包含 `NORMAL`，无异常目标时为空数组 |
| `detections[].box.xyxy` | number[] | 检测框左上角和右下角坐标，格式为 `[x1, y1, x2, y2]` |
| `detections[].box.xywh` | number[] | 检测框中心点和宽高，格式为 `[center_x, center_y, width, height]` |
| `detections[].confidence` | float | YOLO 检测置信度 |
| `detections[].temperature.mean_gray` | float | 检测框区域内灰度均值，范围约 `0` 到 `255` |
| `detections[].temperature.level` | string | 根据灰度均值计算出的异常等级 |
| `elapsed_ms` | float | 本次接口推理耗时，单位毫秒 |

## 5. 异常等级说明

算法会对每个检测框区域计算灰度均值 `mean_gray`，并返回异常等级。`NORMAL` 只作为内部过滤条件，不返回给后端，也不会画框：

| 条件 | 返回值 |
| --- | --- |
| `mean_gray <= 60` | `LOW Lv1` |
| `60 < mean_gray <= 90` | `LOW Lv2` |
| `90 < mean_gray < 120` | `NORMAL`，不返回、不画框 |
| `120 <= mean_gray < 200` | `HIGH Lv2` |
| `mean_gray >= 200` | `HIGH Lv1` |

## 6. 错误响应

### 6.1 图片为空

HTTP 状态码：`400`

```json
{
  "detail": "Uploaded image is empty"
}
```

### 6.2 文件不是有效图片

HTTP 状态码：`400`

```json
{
  "detail": "Uploaded file is not a valid image"
}
```

### 6.3 模型未加载

HTTP 状态码：`503`

```json
{
  "detail": "Model is not loaded"
}
```

### 6.4 参数校验失败

例如 `conf` 小于 `0` 或大于 `1`，FastAPI 会返回 `422`。

```json
{
  "detail": [
    {
      "type": "less_than_equal",
      "loc": ["query", "conf"],
      "msg": "Input should be less than or equal to 1",
      "input": "2",
      "ctx": {
        "le": 1.0
      }
    }
  ]
}
```

## 7. 内网穿透地址配置

每次内网穿透地址变化时，修改 `api/server.py` 文件顶部的变量：

```python
TUNNEL_BASE_URL = "https://你的内网穿透地址"
```

接口返回的标注图地址会按下面格式拼接：

```text
{TUNNEL_BASE_URL}/static/annotated/{filename}.jpg
```

例如：

```text
https://你的内网穿透地址/static/annotated/2f8b5c2f6c0a4b97b7a9f927b7e2c0a1.jpg
```

## 8. 启动方式

在算法服务机器上执行：

```powershell
D:\PythonEnvironment\envs\petroleum_sys\python.exe -m pip install -r requirement.txt
D:\PythonEnvironment\envs\petroleum_sys\python.exe -m uvicorn api.server:app --host 0.0.0.0 --port 8100
```

服务启动后，Java 后端访问：

```text
https://你的内网穿透地址/api/v1/detect
```

## 9. 对接注意事项

- 上传字段名必须为 `image`，否则接口会返回参数错误。
- 当前接口一次请求只处理一张图片。
- 当前接口返回 JSON，标注后的图片通过 `image.url` 访问。
- 坐标单位是像素，坐标基于原始输入图片尺寸。
- `detections` 为空数组表示接口执行成功，但未检测到异常目标；`NORMAL` 结果会被过滤掉。
- 如果 Java 后端需要业务告警，可优先根据 `detections.length`、`confidence` 和 `temperature.level` 判断。
