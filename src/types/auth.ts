import type { JsonObject } from './api'

export type UserRole = 'user' | 'admin' | (string & {})

export interface RegisterParams {
  username: string
  email: string
  password: string
}

export interface LoginParams {
  email: string
  password: string
}

export interface LoginResponseData extends JsonObject {
  token: string
  User_ID: number
  Username: string
  Email: string
  Role: UserRole
  message?: string
}

export interface LoginResponse {
  success: boolean
  data: LoginResponseData
  message?: string
}

export interface RegisterResponseData extends JsonObject {
  User_ID: number
  Username: string
  Email: string
  message: string
}

export interface RegisterResponse {
  success: boolean
  data: RegisterResponseData
  message?: string
}

export interface LogoutResponse {
  success: boolean
  message?: string
}

export interface UserProfileResponse {
  success: boolean
  data: {
    user?: JsonObject
  } & JsonObject
  message?: string
}

export interface UpdateUserResponse {
  success: boolean
  message?: string
}

export interface DeleteUserResponse {
  success: boolean
  message?: string
}

export type VerificationType = 'register' | 'login' | 'reset_password'

export interface SendEmailVerificationParams {
  email: string
  type: VerificationType
}

export interface SendEmailVerificationResponse {
  success: boolean
  message?: string
}

export interface UpdateUserParams {
  username: string
}

export interface UserProfileData {
  user: JsonObject
}

