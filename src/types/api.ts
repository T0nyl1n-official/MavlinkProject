export type JsonObject = Record<string, unknown>

export interface ApiResponse<T extends JsonObject> {
  code: number
  message: string
  /** 部分接口与 mock 返回会带 success 标记 */
  success?: boolean
  data?: T
  token?: string
  expire_time?: number
}

export interface PublicInfoResponse {
  status: 'success'
  message: string
  version: string
}

