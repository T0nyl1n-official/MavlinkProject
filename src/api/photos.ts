import request from './request'

export interface PhotoMessage {
  photo_url: string
  timestamp: string
  lat: number
  lng: number
  h2s_level?: number
  alert?: boolean
}

export const fetchLatestPhoto = async (): Promise<PhotoMessage[]> => {
  try {
    const response = await request.get<PhotoMessage[]>('/output/photos')
    return response.data
  } catch (error) {
    console.error('Failed to fetch photos:', error)
    return []
  }
}

// Mock数据，用于UI演示
export const getMockPhotos = (): PhotoMessage[] => {
  return [
    {
      photo_url: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=drone%20aerial%20view%20of%20industrial%20facility%20with%20gas%20leak%20detection&image_size=landscape_16_9',
      timestamp: new Date().toISOString(),
      lat: 22.5431,
      lng: 114.0523,
      h2s_level: 15.5,
      alert: true
    },
    {
      photo_url: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=drone%20aerial%20view%20of%20factory%20area&image_size=landscape_16_9',
      timestamp: new Date(Date.now() - 60000).toISOString(),
      lat: 22.5451,
      lng: 114.0543,
      h2s_level: 5.2,
      alert: false
    }
  ]
}