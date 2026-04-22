<template>
  <div style="padding: 24px">
    <h1 class="gradient-title">🚁 MAVLink 控制面板</h1>
    <p style="color: var(--text-secondary); margin-bottom: 24px">
      ⚠️ 当前为模拟模式 - 已连接
    </p>

    <!-- 连接状态 -->
    <div style="margin-bottom: 24px; padding: 16px; background: var(--bg-card); border-radius: var(--radius-md); border: 1px solid var(--border-color)">
      <div style="display: flex; align-items: center; gap: 12px">
        <div style="width: 12px; height: 12px; border-radius: 50%; background: var(--success)"></div>
        <span style="font-weight: 500; color: var(--text-primary)">无人机状态：已连接</span>
      </div>
    </div>

    <!-- V1 控制区 -->
    <div style="margin-bottom: 32px">
      <h2 style="font-size: 1.5rem; font-weight: 600; margin-bottom: 16px; color: var(--text-primary)">V1 控制区</h2>
      <div style="background: var(--bg-card); border-radius: var(--radius-md); padding: 24px; border: 1px solid var(--border-color)">
        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 16px; margin-bottom: 24px">
          <!-- 起飞按钮 -->
          <div style="padding: 16px; border-radius: var(--radius-sm); background: var(--bg-body); border: 1px solid var(--border-color)">
            <h3 style="margin-bottom: 12px; font-weight: 500; color: var(--text-primary)">起飞</h3>
            <el-input-number v-model="takeoffAltitude" :min="1" :max="100" label="高度 (m)" style="width: 100%" />
            <el-button type="primary" @click="handleTakeoff" style="width: 100%; margin-top: 12px">
              执行起飞
            </el-button>
          </div>

          <!-- 降落按钮 -->
          <div style="padding: 16px; border-radius: var(--radius-sm); background: var(--bg-body); border: 1px solid var(--border-color)">
            <h3 style="margin-bottom: 12px; font-weight: 500; color: var(--text-primary)">降落</h3>
            <el-input-number v-model="landSpeed" :min="0.1" :max="5" :step="0.1" label="速度 (m/s)" style="width: 100%" />
            <el-button type="primary" @click="handleLand" style="width: 100%; margin-top: 12px">
              执行降落
            </el-button>
          </div>

          <!-- 返航按钮 -->
          <div style="padding: 16px; border-radius: var(--radius-sm); background: var(--bg-body); border: 1px solid var(--border-color)">
            <h3 style="margin-bottom: 12px; font-weight: 500; color: var(--text-primary)">返航</h3>
            <el-button type="primary" @click="handleReturn" style="width: 100%; margin-top: 12px">
              执行返航
            </el-button>
          </div>
        </div>

        <!-- 移动控制 -->
        <div style="padding: 16px; border-radius: var(--radius-sm); background: var(--bg-body); border: 1px solid var(--border-color)">
          <h3 style="margin-bottom: 12px; font-weight: 500; color: var(--text-primary)">移动</h3>
          <div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 12px">
            <el-input v-model="moveLatitude" placeholder="纬度" />
            <el-input v-model="moveLongitude" placeholder="经度" />
            <el-input v-model="moveAltitude" placeholder="高度 (m)" type="number" />
          </div>
          <el-input-number v-model="moveSpeed" :min="1" :max="20" label="速度 (m/s)" style="width: 100%; margin-bottom: 12px" />
          <el-button type="primary" @click="handleMove" style="width: 100%">
            执行移动
          </el-button>
        </div>
      </div>
    </div>

    <!-- V2 控制区 -->
    <div>
      <h2 style="font-size: 1.5rem; font-weight: 600; margin-bottom: 16px; color: var(--text-primary)">V2 控制区</h2>
      <div style="background: var(--bg-card); border-radius: var(--radius-md); padding: 24px; border: 1px solid var(--border-color)">
        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 16px">
          <!-- 一键起飞 -->
          <div style="padding: 16px; border-radius: var(--radius-sm); background: var(--bg-body); border: 1px solid var(--border-color)">
            <h3 style="margin-bottom: 12px; font-weight: 500; color: var(--text-primary)">一键起飞</h3>
            <el-input-number v-model="v2TakeoffAltitude" :min="1" :max="100" label="高度 (m)" style="width: 100%" />
            <el-button type="primary" @click="handleV2Takeoff" style="width: 100%; margin-top: 12px">
              一键起飞
            </el-button>
          </div>

          <!-- 一键降落 -->
          <div style="padding: 16px; border-radius: var(--radius-sm); background: var(--bg-body); border: 1px solid var(--border-color)">
            <h3 style="margin-bottom: 12px; font-weight: 500; color: var(--text-primary)">一键降落</h3>
            <el-input-number v-model="v2LandSpeed" :min="0.1" :max="5" :step="0.1" label="速度 (m/s)" style="width: 100%" />
            <el-button type="primary" @click="handleV2Land" style="width: 100%; margin-top: 12px">
              一键降落
            </el-button>
          </div>

          <!-- 传感器警报模拟 -->
          <div style="padding: 16px; border-radius: var(--radius-sm); background: var(--bg-body); border: 1px solid var(--border-color)">
            <h3 style="margin-bottom: 12px; font-weight: 500; color: var(--text-primary)">传感器警报</h3>
            <el-button type="warning" @click="handleSensorAlert" style="width: 100%">
              发送火灾警报
            </el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

// V1 控制参数
const takeoffAltitude = ref(10)
const landSpeed = ref(1)
const moveLatitude = ref('31.2304')
const moveLongitude = ref('121.4737')
const moveAltitude = ref('20')
const moveSpeed = ref(5)

// V2 控制参数
const v2TakeoffAltitude = ref(10)
const v2LandSpeed = ref(1)

// 处理起飞
const handleTakeoff = () => {
  const params = { altitude: takeoffAltitude.value }
  console.log('Mock模式：发送起飞指令', params)
  ElMessage.info(`Mock模式：已发送起飞指令，高度 ${takeoffAltitude.value}m`)
}

// 处理降落
const handleLand = () => {
  const params = { speed: landSpeed.value }
  console.log('Mock模式：发送降落指令', params)
  ElMessage.info(`Mock模式：已发送降落指令，速度 ${landSpeed.value}m/s`)
}

// 处理移动
const handleMove = () => {
  const params = {
    latitude: parseFloat(moveLatitude.value),
    longitude: parseFloat(moveLongitude.value),
    altitude: parseFloat(moveAltitude.value),
    speed: moveSpeed.value
  }
  console.log('Mock模式：发送移动指令', params)
  ElMessage.info(`Mock模式：已发送移动指令到坐标 (${moveLatitude.value}, ${moveLongitude.value})，高度 ${moveAltitude.value}m`)
}

// 处理返航
const handleReturn = () => {
  console.log('Mock模式：发送返航指令')
  ElMessage.info('Mock模式：已发送返航指令')
}

// 处理 V2 起飞
const handleV2Takeoff = () => {
  const params = { altitude: v2TakeoffAltitude.value }
  console.log('Mock模式：发送V2起飞指令', params)
  ElMessage.info(`Mock模式：已发送V2起飞指令，高度 ${v2TakeoffAltitude.value}m`)
}

// 处理 V2 降落
const handleV2Land = () => {
  const params = { speed: v2LandSpeed.value }
  console.log('Mock模式：发送V2降落指令', params)
  ElMessage.info(`Mock模式：已发送V2降落指令，速度 ${v2LandSpeed.value}m/s`)
}

// 处理传感器警报
const handleSensorAlert = () => {
  const params = {
    sensor_id: 'fire_sensor_001',
    latitude: 31.2304,
    longitude: 121.4737,
    radius: 100,
    photo_count: 5,
    altitude: 50
  }
  console.log('Mock模式：发送传感器警报指令', params)
  ElMessage.info('Mock模式：已发送火灾警报指令')
}
</script>

<style scoped>
.gradient-title {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 16px;
  text-align: center;
  letter-spacing: 2px;
  background: linear-gradient(135deg, var(--primary), var(--primary-light));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
</style>