import request from './request'

export interface PhotoResponse {
  filename: string
  url: string
  timestamp: string
  lat?: number
  lng?: number
}

export interface LatestPhotoResponse {
  photo_url: string
  timestamp: string
  lat?: number
  lng?: number
  h2s_level?: number
  alert?: boolean
}

export const uploadPhoto = async (file: File, droneId?: string): Promise<PhotoResponse> => {
  try {
    const formData = new FormData()
    formData.append('file', file)
    if (droneId) {
      formData.append('drone_id', droneId)
    }

    const response = await request.post<PhotoResponse>('/api/upload/photo', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    return response.data
  } catch (error) {
    console.error('Failed to upload photo:', error)
    throw error
  }
}

export const getLatestPhoto = async (): Promise<LatestPhotoResponse | null> => {
  try {
    const response = await request.get<LatestPhotoResponse>('/api/photo/latest')
    return response.data
  } catch (error) {
    console.error('Failed to get latest photo:', error)
    // 如果接口不存在，返回null
    return null
  }
}

// Mock数据，用于UI演示
export const getMockLatestPhoto = (): LatestPhotoResponse => {
  return {
    photo_url: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=drone%20aerial%20view%20of%20industrial%20facility%20with%20gas%20leak%20detection&image_size=landscape_16_9',
    timestamp: new Date().toISOString(),
    lat: 22.5431,
    lng: 114.0523,
    h2s_level: 15.5,
    alert: true
  }
}