import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            redirect: '/login'
        },
        {
            path: '/login',
            name: 'Login',
            component: () => import('@/views/Login/index.vue')
        },
        {
            path: '/register',
            name: 'Register',
            component: () => import('@/views/Register/index.vue')
        },
        {
            path: '/chain-create',
            name: 'ChainCreate',
            component: () => import('@/views/ChainCreate/index.vue')
        },
        {
            path: '/dashboard',
            name: 'Dashboard',
            component: () => import('@/views/Dashboard/index.vue')
        },
        {
            path: '/chain-manager',
            name: 'ChainManager',
            component: () => import('@/views/ChainManager/index.vue')
        },
        {
            path: '/board',
            name: 'Board',
            component: () => import('@/views/BoardManager/index.vue')
        },
        {
            path: '/mavlink',
            name: 'Mavlink',
            component: () => import('@/views/MavlinkControl/index.vue')
        },
        {
            path: '/monitor',
            name: 'Monitor',
            component: () => import('@/views/Monitor/index.vue')
        },
        {
            path: '/settings',
            name: 'Settings',
            component: () => import('@/views/Settings/index.vue')
        },
        {
            path: '/terminal',
            name: 'Terminal',
            component: () => import('@/views/Terminal/index.vue')
        }
    ]
})

// 路由守卫：未登录用户访问受保护页面时跳转到登录页
router.beforeEach((to, _from, next) => {
    const token = localStorage.getItem('token')

    // 登录页和注册页：如果已经登录则直接跳到默认页
    if (to.path === '/login' || to.path === '/register') {
        if (token) next({ path: '/chain-manager', replace: true })
        else next()
        return
    }

    // 其他页面：未登录则跳转到 /login
    if (!token) next({ path: '/login', replace: true })
    else next()
})

export default router