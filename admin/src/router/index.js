import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/store'

const routes = [
  {
    path: '/install',
    name: 'Install',
    component: () => import('@/views/install/Index.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/Register.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/404.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/dashboard'
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/Index.vue')
      },
      {
        path: 'devices',
        name: 'DeviceList',
        component: () => import('@/views/device/List.vue')
      },
      {
        path: 'devices/import',
        name: 'DeviceImport',
        component: () => import('@/views/device/DeviceImport.vue')
      },
      {
        path: 'devices/:id',
        name: 'DeviceDetail',
        component: () => import('@/views/device/Detail.vue'),
        props: true
      },
      {
        path: 'device-map',
        name: 'DeviceMap',
        component: () => import('@/views/device/Map.vue')
      },
      {
        path: 'groups',
        name: 'GroupList',
        component: () => import('@/views/device/Group.vue')
      },
      {
        path: 'device-events',
        name: 'DeviceEventList',
        component: () => import('@/views/device/Events.vue')
      },
      {
        path: 'device-logs',
        name: 'DeviceLogList',
        component: () => import('@/views/device/Logs.vue')
      },
      {
        path: 'rules',
        name: 'RuleList',
        component: () => import('@/views/rule/List.vue')
      },
      {
        path: 'families',
        name: 'FamilyList',
        component: () => import('@/views/family/List.vue')
      },
      {
        path: 'scenes',
        name: 'SceneList',
        component: () => import('@/views/scene/List.vue')
      },
      {
        path: 'tasks',
        name: 'TaskList',
        component: () => import('@/views/task/List.vue')
      },
      {
        path: 'models',
        name: 'ModelList',
        component: () => import('@/views/model/List.vue')
      },
      {
        path: 'models/:id',
        name: 'ModelDetail',
        component: () => import('@/views/model/Detail.vue'),
        props: true
      },
      {
        path: 'models/:id/designer',
        name: 'ModelDesigner',
        component: () => import('@/views/model/Designer.vue'),
        props: true
      },
      {
        path: 'gateway/servers',
        name: 'GatewayServers',
        component: () => import('@/views/modbus/Server.vue')
      },
      {
        path: 'gateway/connections',
        name: 'GatewayConnectionsAll',
        component: () => import('@/views/modbus/Connections.vue')
      },
      {
        path: 'gateway/servers/:id/connections',
        name: 'GatewayConnections',
        component: () => import('@/views/modbus/Connections.vue')
      },
      {
        path: 'gateway/servers/:serverId/connections/:connId',
        name: 'GatewayConnectionDetail',
        component: () => import('@/views/modbus/ConnectionDetail.vue'),
        props: true
      },
      {
        path: 'scada',
        name: 'ScadaList',
        component: () => import('@/views/scada/List.vue')
      },
      {
        path: 'scada/:id/editor',
        name: 'ScadaEditor',
        component: () => import('@/views/scada/Editor.vue'),
        props: true
      },
      {
        path: 'scada/:id/preview',
        name: 'ScadaPreview',
        component: () => import('@/views/scada/Preview.vue'),
        props: true
      },
      {
        path: 'components',
        name: 'ComponentLibrary',
        component: () => import('@/views/scada/ComponentLibrary.vue')
      },
      {
        path: 'ota/firmware',
        name: 'FirmwareList',
        component: () => import('@/views/ota/FirmwareList.vue')
      },
      {
        path: 'ota/logs',
        name: 'UpgradeLogs',
        component: () => import('@/views/ota/UpgradeLogs.vue')
      },
      {
        path: 'maintenance/plans',
        name: 'MaintenancePlans',
        component: () => import('@/views/maintenance/PlanList.vue'),
        meta: { title: '维保计划' }
      },
      {
        path: 'maintenance/orders',
        name: 'MaintenanceOrders',
        component: () => import('@/views/maintenance/OrderList.vue'),
        meta: { title: '维保工单' }
      },
      {
        path: 'maintenance/orders/:id',
        name: 'MaintenanceOrderDetail',
        component: () => import('@/views/maintenance/OrderDetail.vue'),
        props: true,
        meta: { title: '工单详情' }
      },
      {
        path: 'maintenance/stats',
        name: 'MaintenanceStats',
        component: () => import('@/views/maintenance/Stats.vue'),
        meta: { title: '维保统计' }
      },
      {
        path: 'accounts',
        name: 'AccountList',
        component: () => import('@/views/account/List.vue')
      },
      {
        path: 'accounts/:id',
        name: 'AccountDetail',
        component: () => import('@/views/account/Detail.vue'),
        props: true
      },
      {
        path: 'orgs',
        name: 'OrgList',
        component: () => import('@/views/org/List.vue')
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/mine/Profile.vue')
      },
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 安装模式检测：master.yaml 缺失时后端仅提供 /api/setup/*，前端必须进入安装向导。
// 正常模式下该接口不存在（请求失败），按未安装处理为 false；检测结果缓存，只请求一次。
let installChecked = false
let installing = false

async function detectInstallMode() {
  if (installChecked) return installing
  installChecked = true
  try {
    const res = await fetch('/api/setup/status')
    const data = await res.json()
    installing = data?.installed === false
  } catch (e) {
    installing = false
  }
  return installing
}

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()

  if (to.path !== '/install' && (await detectInstallMode())) {
    next('/install')
    return
  }

  if (to.meta.requiresAuth !== false && !userStore.token) {
    next('/login')
    return
  }

  if (to.path === '/login' && userStore.token) {
    next('/dashboard')
    return
  }

  next()
})

export default router