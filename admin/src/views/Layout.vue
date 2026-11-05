<template>
  <div class="layout-container">
    <!-- 侧边栏 -->
    <aside class="sidebar" :class="{ collapsed: isCollapsed }">
      <div class="logo">
          <div class="logo-icon">
            <img v-if="logoSrc" :src="logoSrc" alt="logo" class="logo-img" />
            <el-icon v-else :size="32"><Connection /></el-icon>
          </div>
          <transition name="fade">
            <span v-if="!isCollapsed" class="logo-text">{{ oemStore.getSystemName }}</span>
          </transition>
        </div>

      <el-menu
        :default-active="currentRoute"
        :collapse="isCollapsed"
        :router="true"
        class="sidebar-menu"
      >
        <!-- 设备 -->
        <el-sub-menu index="device" class="menu-group">
          <template #title>
            <el-icon><Cpu /></el-icon>
            <span>设备</span>
          </template>
          <el-menu-item index="/devices">
            <el-icon><Cpu /></el-icon>
            <span>所有设备</span>
          </el-menu-item>
          <el-menu-item index="/device-events">
            <el-icon><Bell /></el-icon>
            <span>异常报警</span>
          </el-menu-item>
          <el-menu-item index="/device-logs">
            <el-icon><Tickets /></el-icon>
            <span>设备日志</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 物模型 -->
        <el-sub-menu index="model" class="menu-group">
          <template #title>
            <el-icon><Operation /></el-icon>
            <span>物模型</span>
          </template>
          <el-menu-item index="/models">
            <el-icon><Notebook /></el-icon>
            <span>所有物模型</span>
          </el-menu-item>
          <el-menu-item index="/rules">
            <el-icon><MagicStick /></el-icon>
            <span>规则引擎</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 智能家居 -->
        <el-sub-menu index="smarthome" class="menu-group">
          <template #title>
            <el-icon><HomeFilled /></el-icon>
            <span>智能家居</span>
          </template>
          <el-menu-item index="/families">
            <el-icon><House /></el-icon>
            <span>家庭管理</span>
          </el-menu-item>
          <el-menu-item index="/scenes">
            <el-icon><Sunny /></el-icon>
            <span>智能场景</span>
          </el-menu-item>
          <el-menu-item index="/tasks">
            <el-icon><Timer /></el-icon>
            <span>定时任务</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 组态大屏 -->
        <el-sub-menu index="scada" class="menu-group">
          <template #title>
            <el-icon><Monitor /></el-icon>
            <span>组态大屏</span>
          </template>
          <el-menu-item index="/scada">
            <el-icon><Histogram /></el-icon>
            <span>我的组态</span>
          </el-menu-item>
          <el-menu-item index="/components">
            <el-icon><FolderOpened /></el-icon>
            <span>组件库</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 网关服务 -->
        <el-sub-menu index="gateway" class="menu-group">
          <template #title>
            <el-icon><Connection /></el-icon>
            <span>网关服务</span>
          </template>
          <el-menu-item index="/gateway/servers">
            <el-icon><Monitor /></el-icon>
            <span>服务器</span>
          </el-menu-item>
          <el-menu-item index="/gateway/connections">
            <el-icon><Link /></el-icon>
            <span>连接</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- OTA 升级 -->
        <el-sub-menu index="ota" class="menu-group">
          <template #title>
            <el-icon><UploadFilled /></el-icon>
            <span>OTA升级</span>
          </template>
          <el-menu-item index="/ota/firmware">
            <el-icon><Box /></el-icon>
            <span>固件管理</span>
          </el-menu-item>
          <el-menu-item index="/ota/logs">
            <el-icon><Clock /></el-icon>
            <span>升级日志</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 维保管理 -->
        <el-sub-menu index="maintenance" class="menu-group">
          <template #title>
            <el-icon><Tools /></el-icon>
            <span>维保管理</span>
          </template>
          <el-menu-item index="/maintenance/plans">
            <el-icon><Calendar /></el-icon>
            <span>维保计划</span>
          </el-menu-item>
          <el-menu-item index="/maintenance/orders">
            <el-icon><Tickets /></el-icon>
            <span>维保工单</span>
          </el-menu-item>
          <el-menu-item index="/maintenance/stats">
            <el-icon><DataAnalysis /></el-icon>
            <span>维保统计</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 用户管理 -->
        <el-sub-menu index="account" class="menu-group">
          <template #title>
            <el-icon><UserFilled /></el-icon>
            <span>用户管理</span>
          </template>
          <el-menu-item index="/accounts">
            <el-icon><User /></el-icon>
            <span>所有用户</span>
          </el-menu-item>
          <el-menu-item index="/orgs">
            <el-icon><OfficeBuilding /></el-icon>
            <span>组织管理</span>
          </el-menu-item>
          <el-menu-item index="/profile">
            <el-icon><Avatar /></el-icon>
            <span>个人中心</span>
          </el-menu-item>
        </el-sub-menu>
      </el-menu>
    </aside>

    <!-- 主内容区 -->
    <div class="main-wrapper">
      <!-- 顶部导航 -->
      <header class="topbar">
        <div class="left">
          <el-icon class="collapse-btn" @click="isCollapsed = !isCollapsed">
            <Fold v-if="!isCollapsed" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/" class="breadcrumb">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentRouteName">{{ currentRouteName }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="right">
          <div class="topbar-actions">
            <el-tooltip content="消息通知" placement="bottom">
              <el-icon class="action-icon"><Bell /></el-icon>
            </el-tooltip>
            <el-tooltip content="全屏" placement="bottom">
              <el-icon class="action-icon"><FullScreen /></el-icon>
            </el-tooltip>

          </div>
          <el-dropdown @command="handleCommand" trigger="click">
            <span class="user-info">
              <el-avatar :size="36" class="user-avatar" :src="avatarUrl">
                {{ userStore.username?.[0]?.toUpperCase() || 'U' }}
              </el-avatar>
              <div class="user-detail">
                <span class="username">{{ userStore.username || '用户' }}</span>
                <span class="role" v-if="userStore.isAdmin">管理员</span>
              </div>
              <el-icon class="arrow"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>个人中心
                </el-dropdown-item>
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>设置
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- 内容区 -->
      <main class="content">
        <router-view v-slot="{ Component }">
          <transition name="fade-slide" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>

    <!-- 窗口管理系统 -->
    <WindowManager ref="windowManagerRef" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, provide, watch } from 'vue'

defineOptions({
  name: 'Layout'
})
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import {
  Connection, DataAnalysis, MapLocation,
  Notebook, Monitor, Notification, Shop, User, UserFilled, Avatar,
  FolderOpened, Setting, Grid, Fold, Expand,
  ArrowDown, Bell, FullScreen, SwitchButton,
  Cpu, MagicStick, HomeFilled, Sunny, Timer, Operation, Histogram, House, Tickets, Link, OfficeBuilding,
  UploadFilled, Box, Clock, Tools, Calendar
} from '@element-plus/icons-vue'
import { useUserStore, useOEMStore } from '@/store'
import { getMe } from '@/api/user'
import WindowManager from '@/components/WindowManager/index.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const oemStore = useOEMStore()
const windowManagerRef = ref(null)

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const avatarUrl = computed(() => {
  if (!userStore.avatar) return ''
  if (userStore.avatar.startsWith('http')) return userStore.avatar
  return baseUrl + userStore.avatar
})

// Logo 图片：OEM 配置优先，空值时显示默认 Connection 图标
const logoSrc = computed(() => {
  return oemStore.getLogo || ''
})

const isCollapsed = ref(localStorage.getItem('sidebar_collapsed') === 'true')

onMounted(async () => {
  if (userStore.isLoggedIn && !userStore.id) {
    try {
      const res = await getMe()
      userStore.setUser(res.data)
    } catch (e) {
      console.error('获取用户信息失败', e)
    }
  }
})

watch(isCollapsed, (val) => {
  localStorage.setItem('sidebar_collapsed', String(val))
})

router.afterEach(() => {
  windowManagerRef.value?.minimizeAll()
})

const currentRoute = computed(() => route.path)

const currentRouteName = computed(() => {
  const map = {
    '/dashboard': '控制台',
    '/devices': '所有设备',
    '/groups': '分组管理',
    '/device-events': '异常报警',
    '/device-logs': '设备日志',
    '/rules': '规则引擎',
    '/scenes': '智能场景',
    '/families': '家庭管理',
    '/tasks': '定时任务',
    '/models': '所有物模型',
    '/scada': '我的组态',
    '/components': '组件库',
    '/gateway/servers': '网关服务器',
    '/gateway/connections': '连接',
    '/ota/firmware': '固件管理',
    '/ota/logs': '升级日志',
    '/maintenance/plans': '维保计划',
    '/maintenance/orders': '维保工单',
    '/maintenance/stats': '维保统计',
    '/accounts': '所有用户',
    '/orgs': '组织管理',
    '/profile': '个人中心'
  }
  if (route.path.startsWith('/maintenance/orders/')) return '工单详情'
  return map[route.path]
})

const handleCommand = (command) => {
  switch (command) {
    case 'logout':
      ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        userStore.logout()
        router.push('/login')
      }).catch(() => {})
      break
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      break
  }
}

provide('openWindow', (app) => windowManagerRef.value?.openWindow(app))

defineExpose({
  openWindow: (app) => windowManagerRef.value?.openWindow(app)
})
</script>

<style lang="scss" scoped>
.layout-container {
  display: flex;
  width: 100%;
  height: 100vh;
  background: #f5f7fa;
}

.sidebar {
  width: 240px;
  height: 100vh;
  background: linear-gradient(180deg, #1a1a2e 0%, #16213e 100%);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  box-shadow: 4px 0 20px rgba(0, 0, 0, 0.15);

  &.collapsed {
    width: 68px;
  }

  .logo {
    height: 70px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 0 16px;
    background: rgba(255, 255, 255, 0.03);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);

    .logo-icon {
      width: 40px;
      height: 40px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, #00d9ff 0%, #0099ff 100%);
      border-radius: 10px;
      color: #fff;
      box-shadow: 0 4px 15px rgba(0, 217, 255, 0.3);

      .logo-img {
        width: 100%;
        height: 100%;
        object-fit: contain;
        border-radius: 6px;
        background: #fff;
      }
    }

    .logo-text {
      font-size: 18px;
      font-weight: 700;
      color: #fff;
      letter-spacing: 2px;
      white-space: nowrap;
      text-shadow: 0 2px 10px rgba(0, 217, 255, 0.5);
    }
  }

  .sidebar-menu {
    background: transparent;
    border: none;
    padding: 12px 0;

    :deep(.el-menu-item),
    :deep(.el-sub-menu__title) {
      height: 52px;
      line-height: 52px;
      color: rgba(255, 255, 255, 0.7);
      //transition: all 0.3s ease;

      .el-icon {
        font-size: 18px;
        margin-right: 12px;
      }

      &:hover {
        background: rgba(0, 217, 255, 0.1);
        color: #00d9ff;
      }
    }

    :deep(.el-sub-menu__title) {
      font-size: 16px;
      font-weight: 600;
    }

    :deep(.el-menu-item.is-active) {
      background: linear-gradient(90deg, rgba(0, 217, 255, 0.2) 0%, rgba(0, 153, 255, 0.1) 100%);
      color: #00d9ff;
      border-left: 3px solid #00d9ff;
      margin-left: 9px;
      padding-left: 12px;
    }

    :deep(.el-sub-menu) {
      .el-menu {
        background: transparent;
        
        .el-menu-item {
          height: 46px;
          line-height: 46px;
          margin: 2px 12px;
          padding-left: 13px !important;
          font-size: 13px;

          &.is-active {
            background: linear-gradient(90deg, rgba(0, 217, 255, 0.15) 0%, rgba(0, 153, 255, 0.05) 100%);
            color: #00d9ff;
          }
        }
      }

      .el-sub-menu__title {
        &:hover {
          background: rgba(0, 217, 255, 0.1);
        }
      }
    }
  }
}

.main-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.topbar {
  height: 70px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: #fff;
  box-shadow: 0 2px 20px rgba(0, 0, 0, 0.05);
  position: relative;
  z-index: 10;

  .left {
    display: flex;
    align-items: center;
    gap: 20px;

    .collapse-btn {
      font-size: 40px;
      color: #666;
      cursor: pointer;
      padding: 8px;
      border-radius: 8px;
      transition: all 0.3s;

      &:hover {
        background: #f5f7fa;
        color: #00d9ff;
      }
    }

    .breadcrumb {
      :deep(.el-breadcrumb__item) {
        .el-breadcrumb__inner {
          color: #666;
          font-weight: 500;
        }
        &:last-child .el-breadcrumb__inner {
          color: #1a1a2e;
          font-weight: 600;
        }
      }
    }
  }

  .right {
    display: flex;
    align-items: center;
    gap: 20px;

    .topbar-actions {
      display: flex;
      gap: 8px;

      .action-icon {
        font-size: 20px;
        color: #666;
        padding: 10px;
        border-radius: 10px;
        cursor: pointer;
        transition: all 0.3s;

        &:hover {
          background: #f5f7fa;
          color: #00d9ff;
        }
      }


    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 12px;
      cursor: pointer;
      padding: 6px 12px 6px 6px;
      border-radius: 12px;
      transition: all 0.3s;

      &:hover {
        background: #f5f7fa;
      }

      .user-avatar {
        background: linear-gradient(135deg, #00d9ff 0%, #0099ff 100%);
        color: #fff;
        font-weight: 600;
      }

      .user-detail {
        display: flex;
        flex-direction: column;

        .username {
          font-size: 14px;
          font-weight: 600;
          color: #1a1a2e;
        }

        .role {
          font-size: 12px;
          color: #999;
        }
      }

      .arrow {
        color: #999;
      }
    }
  }
}

.content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  overflow-x: hidden;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}
.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}
</style>