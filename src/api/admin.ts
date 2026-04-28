import request from './request'
import { config } from '@/utils/mockService'

export function getAllUsersApi() {
    if (!config.USE_REAL_API) {
        return Promise.resolve({
            code: 0,
            success: true,
            data: {
                users: [
                    { User_ID: 1, Username: 'admin', Email: 'admin@example.com', Role: 'admin' },
                    { User_ID: 2, Username: 'user1', Email: 'user1@example.com', Role: 'user' },
                    { User_ID: 3, Username: 'user2', Email: 'user2@example.com', Role: 'user' }
                ]
            },
            message: 'Success'
        })
    }
    return request.get('/admin/all-profile').then((res: any) => ({
        code: typeof res?.code === 'number' ? res.code : 0,
        success: res?.success ?? res?.code === 0,
        message: res?.message || 'Success',
        data: {
            users: Array.isArray(res?.data)
                ? res.data.map((user: any) => ({
                    User_ID: Number(user.user_id || 0),
                    Username: user.username || '',
                    Email: user.email || '',
                    Role: user.is_admin ? 'admin' : 'user'
                }))
                : (res?.data?.users || [])
        },
        pagination: res?.pagination
    }))
}

export function deleteUserApi(userId: number) {
    if (!config.USE_REAL_API) {
        return Promise.resolve({
            code: 0,
            success: true,
            message: 'User deleted successfully'
        })
    }
    return request.post('/admin/delete-user', { userId })
}

export function updateUserRoleApi(userId: number, role: string) {
    if (!config.USE_REAL_API) {
        return Promise.resolve({
            code: 0,
            success: true,
            message: 'User role updated successfully'
        })
    }
    return request.post('/admin/update-role', { userId, role })
}
