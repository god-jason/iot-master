import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'
import { registerFilters } from './utils/format'
import { initOEM } from './utils/oem'
import './styles/common.scss'
import './iotfont/iotfont.css'

const app = createApp(App)

registerFilters(app)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })

// 初始化 OEM 定制化配置（先加载缓存，再异步拉取最新）
initOEM()

app.mount('#app')
