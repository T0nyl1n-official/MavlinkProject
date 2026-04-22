import request from '@/api/request'
import { mockService, config } from '@/utils/mockService'
import type {
  DeleteUserResponse,
  LoginParams,
  LoginResponse,
  LogoutResponse,
  RegisterParams,
  RegisterResponse,
  SendEmailVerificationParams,
  SendEmailVerificationResponse,
  UpdateUserParams,
  UpdateUserResponse,
  UserProfileResponse
} from '@/types/auth'

export function registerApi(params: RegisterParams): Promise<RegisterResponse> {
  if (!config.USE_REAL_API) {
    return mockService.register(params.username, params.email, params.password).then(user => ({
      code: 0,
      success: true,
      data: {
        User_ID: user.User_ID,
        Username: user.Username,
        Email: user.Email,
        message: 'Registration successful'
      },
      message: 'Registration successful'
    }))
  }
  return request.post('/users/register', params)
}

export function loginApi(params: LoginParams): Promise<LoginResponse> {
  if (!config.USE_REAL_API) {
    return mockService.login(params.email, params.password).then(user => ({
      code: 0,
      success: true,
      data: {
        token: user.token,
        User_ID: user.User_ID,
        Username: user.Username,
        Email: user.Email,
        Role: user.Role
      },
      message: 'Login successful'
    }))
  }
  return request.post('/users/login', params)
}

export function logoutApi(): Promise<LogoutResponse> {
  if (!config.USE_REAL_API) {
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Logout successful'
    })
  }
  return request.post('/users/logout')
}

export function getProfileApi(): Promise<UserProfileResponse> {
  if (!config.USE_REAL_API) {
    return mockService.getProfile().then(user => ({
      code: 0,
      success: true,
      data: {
        user: {
          User_ID: user.User_ID,
          Username: user.Username,
          Email: user.Email,
          Role: user.Role
        }
      },
      message: 'Profile retrieved successfully'
    }))
  }
  return request.get('/users/profile')
}

export function updateProfileApi(params: UpdateUserParams): Promise<UpdateUserResponse> {
  if (!config.USE_REAL_API) {
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Profile updated successfully'
    })
  }
  return request.post('/users/update', params)
}

export function deleteUserApi(): Promise<DeleteUserResponse> {
  if (!config.USE_REAL_API) {
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'User deleted successfully'
    })
  }
  return request.post('/users/delete')
}

export function sendEmailVerificationApi(
  params: SendEmailVerificationParams
): Promise<SendEmailVerificationResponse> {
  if (!config.USE_REAL_API) {
    return Promise.resolve({
      code: 0,
      success: true,
      message: 'Verification email sent'
    })
  }
  return request.post('/users/send-email-verification', params)
}