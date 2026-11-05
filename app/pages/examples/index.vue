<template>
  <view class="page-examples-index">
    <DsNavBar title="组件示例" />
    
    <view class="examples-content">
      <DsSection title="DsIcon 图标" type="card">
        <view class="example-grid">
          <view class="example-item">
            <DsIcon name="home" :size="20" color="#333" />
            <text>20px</text>
          </view>
          <view class="example-item">
            <DsIcon name="home" :size="24" color="#5BA3E8" />
            <text>24px</text>
          </view>
          <view class="example-item">
            <DsIcon name="home" :size="32" color="#67c23a" />
            <text>32px</text>
          </view>
          <view class="example-item">
            <DsIcon name="home" :size="40" color="#f56c6c" />
            <text>40px</text>
          </view>
        </view>
        <view class="example-grid">
          <view class="example-item">
            <DsIcon name="setup" :size="24" />
            <text>setup</text>
          </view>
          <view class="example-item">
            <DsIcon name="my" :size="24" />
            <text>member</text>
          </view>
          <view class="example-item">
            <DsIcon name="conversation" :size="24" />
            <text>message</text>
          </view>
          <view class="example-item">
            <DsIcon name="store" :size="24" />
            <text>store</text>
          </view>
          <view class="example-item">
            <DsIcon name="list" :size="24" />
            <text>order</text>
          </view>
          <view class="example-item">
            <DsIcon name="wallet" :size="24" />
            <text>wallet</text>
          </view>
        </view>
        <view class="example-grid">
          <view class="example-item">
            <DsIcon name="back" :size="24" />
            <text>back</text>
          </view>
          <view class="example-item">
            <DsIcon name="next" :size="24" />
            <text>next</text>
          </view>
          <view class="example-item">
            <DsIcon name="down_arrow" :size="24" />
            <text>down_arrow</text>
          </view>
          <view class="example-item">
            <DsIcon name="up_arrow" :size="24" />
            <text>up_arrow</text>
          </view>
          <view class="example-item">
            <DsIcon name="affirm" :size="24" />
            <text>affirm</text>
          </view>
          <view class="example-item">
            <DsIcon name="warning" :size="24" />
            <text>warning</text>
          </view>
        </view>
      </DsSection>

      <DsSection title="DsAvatar 头像" type="card">
        <view class="example-grid">
          <view class="example-item">
            <DsAvatar icon="my" size="mini" bgColor="#5BA3E8" />
            <text>mini</text>
          </view>
          <view class="example-item">
            <DsAvatar icon="my" size="small" bgColor="#67c23a" />
            <text>small</text>
          </view>
          <view class="example-item">
            <DsAvatar icon="my" size="medium" bgColor="#e6a23c" />
            <text>medium</text>
          </view>
          <view class="example-item">
            <DsAvatar icon="my" size="large" bgColor="#f56c6c" />
            <text>large</text>
          </view>
          <view class="example-item">
            <DsAvatar icon="my" size="xlarge" bgColor="#909399" />
            <text>xlarge</text>
          </view>
        </view>
        <view class="example-grid">
          <view class="example-item">
            <DsAvatar icon="my" shape="circle" size="medium" />
            <text>圆形</text>
          </view>
          <view class="example-item">
            <DsAvatar icon="my" shape="square" size="medium" />
            <text>方形</text>
          </view>
          <view class="example-item">
            <DsAvatar icon="my" size="mini" :badge="{ type: 'dot' }" />
            <text>迷你红点</text>
          </view>
          <view class="example-item">
            <DsAvatar icon="my" size="small" :badge="{ type: 'dot' }" />
            <text>小型红点</text>
          </view>
          <view class="example-item">
            <DsAvatar icon="my" size="medium" :badge="{ type: 'text', value: '3' }" />
            <text>数字徽标</text>
          </view>
        </view>
      </DsSection>

      <DsSection title="DsBadge 徽标" type="card">
        <view class="example-row">
          <DsBadge :value="1">
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
          <DsBadge :value="9">
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
          <DsBadge :value="99">
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
          <DsBadge :value="100">
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
          <DsBadge dot>
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
        </view>
        <view class="example-row">
          <DsBadge :value="5" type="success">
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
          <DsBadge :value="5" type="warning">
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
          <DsBadge :value="5" type="info">
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
        </view>
        <view class="example-row">
          <DsBadge :value="5" position="top-left">
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
          <DsBadge :value="5" position="bottom-left">
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
          <DsBadge :value="5" position="bottom-right">
            <DsIcon name="conversation" :size="32" />
          </DsBadge>
        </view>
      </DsSection>

      <DsSection title="DsButton 按钮" type="card">
        <view class="example-row">
          <DsButton text="主要" type="primary" />
          <DsButton text="成功" type="success" />
          <DsButton text="警告" type="warning" />
          <DsButton text="危险" type="danger" />
          <DsButton text="默认" type="default" />
          <DsButton text="幽灵" type="ghost" />
        </view>
        <view class="example-row">
          <DsButton text="迷你" size="mini" />
          <DsButton text="小型" size="small" />
          <DsButton text="中型" size="medium" />
          <DsButton text="大型" size="large" />
        </view>
        <view class="example-row">
          <DsButton icon="add" text="带图标" />
          <DsButton icon="search" text="搜索" />
          <DsButton loading text="加载中" />
          <DsButton disabled text="禁用" />
          <DsButton text="圆角" round />
        </view>
        <view class="example-row">
          <DsButton text="仅图标" icon="add" plain />
          <DsButton text="自定义颜色" bgColor="#909399" color="#fff" />
          <DsButton text="边框按钮" border />
        </view>
      </DsSection>

      <DsSection title="DsSwitch 开关" type="card">
        <view class="example-row">
          <DsSwitch :value="switchValue" label="默认开启" @change="(val) => switchValue = val" />
          <DsSwitch :value="!switchValue" label="默认关闭" @change="(val) => switchValue = !val" />
        </view>
        <view class="example-row">
          <DsSwitch :value="true" activeColor="#67c23a" />
          <DsSwitch :value="true" activeColor="#f56c6c" />
          <DsSwitch :value="true" inactiveColor="#ccc" />
          <DsSwitch :value="false" disabled />
        </view>
      </DsSection>

      <DsSection title="DsInput 输入框" type="card">
        <DsFormItem label="基础输入">
          <DsInput v-model="inputValue" placeholder="请输入内容" clearable />
        </DsFormItem>
        <DsFormItem label="密码输入">
          <DsInput v-model="passwordValue" type="password" placeholder="请输入密码" />
        </DsFormItem>
        <DsFormItem label="带图标">
          <DsInput v-model="inputValue" placeholder="请输入" prefixIcon="phone" suffixIcon="search" />
        </DsFormItem>
        <DsFormItem label="错误状态">
          <DsInput v-model="inputValue" placeholder="请输入" :error="true" errorMessage="输入格式错误" />
        </DsFormItem>
        <DsFormItem label="帮助文字">
          <DsInput v-model="inputValue" placeholder="请输入" helpText="请输入正确的格式" />
        </DsFormItem>
        <DsFormItem label="禁用状态">
          <DsInput v-model="inputValue" placeholder="禁用状态" disabled />
        </DsFormItem>
      </DsSection>

      <DsSection title="InputBar 输入条" type="card">
        <DsInputBar placeholder="搜索商品" />
        <DsInputBar placeholder="输入手机号" prefixIcon="phone" />
        <DsInputBar placeholder="输入邮箱" prefixIcon="mail" />
      </DsSection>

      <DsSection title="DsNumberInput 数字输入" type="card">
        <DsFormItem label="基础用法">
          <DsNumberInput v-model="numberValue" :min="1" :max="100" unit="件" />
        </DsFormItem>
        <DsFormItem label="带精度">
          <DsNumberInput v-model="decimalValue" :min="0" :max="10" :precision="2" unit="kg" />
        </DsFormItem>
        <DsFormItem label="步长为5">
          <DsNumberInput v-model="stepValue" :min="0" :max="100" :step="5" />
        </DsFormItem>
        <DsFormItem label="禁用状态">
          <DsNumberInput v-model="numberValue" :min="1" :max="100" disabled />
        </DsFormItem>
      </DsSection>

      <DsSection title="DsComboBox 下拉选择" type="card">
        <DsFormItem label="基础选择">
          <DsComboBox v-model="comboValue" :options="comboOptions" />
        </DsFormItem>
        <DsFormItem label="带占位符">
          <DsComboBox v-model="comboValue2" :options="comboOptions" placeholder="请选择选项" />
        </DsFormItem>
        <DsFormItem label="已选择">
          <DsComboBox v-model="comboValue3" :options="comboOptions" />
        </DsFormItem>
      </DsSection>

      <DsSection title="Radio 单选框" type="card">
        <DsFormItem label="基础单选">
          <DsRadio v-model="radioValue" :options="radioOptions" />
        </DsFormItem>
        <DsFormItem label="带图标">
          <DsRadio v-model="radioValue2" :options="radioOptionsWithIcon" />
        </DsFormItem>
        <DsFormItem label="横向排列">
          <DsRadio v-model="radioValue3" :options="radioOptions" direction="horizontal" />
        </DsFormItem>
      </DsSection>

      <DsSection title="Checkbox 复选框" type="card">
        <DsFormItem label="基础多选">
          <DsCheckbox v-model="checkboxValue" :options="checkboxOptions" />
        </DsFormItem>
        <DsFormItem label="带图标">
          <DsCheckbox v-model="checkboxValue2" :options="checkboxOptionsWithIcon" />
        </DsFormItem>
        <DsFormItem label="横向排列">
          <DsCheckbox v-model="checkboxValue3" :options="checkboxOptions" direction="horizontal" />
        </DsFormItem>
      </DsSection>

      <DsSection title="DsList & DsListItem 列表" type="card">
        <DsList>
          <DsListItem title="设置" desc="系统设置" icon="setup" arrow clickable />
          <DsListItem title="通知" desc="消息通知" icon="conversation" arrow clickable />
          <DsListItem title="帮助" desc="帮助中心" icon="help-circle" arrow clickable />
        </DsList>
        <DsList bordered>
          <DsListItem title="带标签" icon="store" :tag="{ text: '推荐', type: 'primary' }" arrow />
          <DsListItem title="带头像" avatar="/static/logo.png" desc="头像示例" arrow />
        </DsList>
        <DsList striped>
          <DsListItem title="条纹列表1" desc="第一项" />
          <DsListItem title="条纹列表2" desc="第二项" />
          <DsListItem title="条纹列表3" desc="第三项" />
        </DsList>
        <DsList>
          <DsListItem title="带额外内容" desc="描述信息" extra="右侧内容" arrow />
          <DsListItem title="自定义内容" arrow>
            <template #content>
              <text style="color: #67c23a;">自定义内容</text>
            </template>
          </DsListItem>
        </DsList>
      </DsSection>

      <DsSection title="Grid 网格" type="card">
        <DsGrid :items="gridItems" :columns="4" />
        <DsGrid :items="gridItems2" :columns="3" />
        <DsGrid :items="gridItems3" :columns="2" bordered />
      </DsSection>

      <DsSection title="NavBar 导航栏" type="card">
        <DsNavBar title="导航标题" />
        <DsNavBar title="透明导航" transparent />
        <DsNavBar title="自定义颜色" color="#5BA3E8" />
        <DsNavBar :showBack="false" title="无返回">
          <template #right>
            <DsIcon name="setup" :size="24" />
          </template>
        </DsNavBar>
      </DsSection>

      <DsSection title="DsTabs 选项卡" type="card">
        <DsTabs v-model="activeTab" type="line">
          <DsTabItem label="全部" :index="0" :active="activeTab === 0" @click="activeTab = 0" />
          <DsTabItem label="待处理" :index="1" :active="activeTab === 1" @click="activeTab = 1" />
          <DsTabItem label="已完成" :index="2" :active="activeTab === 2" @click="activeTab = 2" />
        </DsTabs>
        <DsTabs v-model="activeTab2" type="line">
          <DsTabItem label="全部" :index="0" :active="activeTab2 === 0" @click="activeTab2 = 0" badge="12" />
          <DsTabItem label="待处理" :index="1" :active="activeTab2 === 1" @click="activeTab2 = 1" badge="3" />
          <DsTabItem label="已完成" :index="2" :active="activeTab2 === 2" @click="activeTab2 = 2" />
        </DsTabs>
        <DsTabs v-model="activeTab3" type="card">
          <DsTabItem label="选项1" :index="0" :active="activeTab3 === 0" @click="activeTab3 = 0" />
          <DsTabItem label="选项2" :index="1" :active="activeTab3 === 1" @click="activeTab3 = 1" />
          <DsTabItem label="选项3" :index="2" :active="activeTab3 === 2" @click="activeTab3 = 2" />
        </DsTabs>
        <DsTabs v-model="activeTab4" scrollable type="line">
          <DsTabItem label="标签很长很长很长很长" :index="0" :active="activeTab4 === 0" @click="activeTab4 = 0" />
          <DsTabItem label="标签二" :index="1" :active="activeTab4 === 1" @click="activeTab4 = 1" />
          <DsTabItem label="标签三" :index="2" :active="activeTab4 === 2" @click="activeTab4 = 2" />
          <DsTabItem label="标签四" :index="3" :active="activeTab4 === 3" @click="activeTab4 = 3" />
          <DsTabItem label="标签五" :index="4" :active="activeTab4 === 4" @click="activeTab4 = 4" />
        </DsTabs>
      </DsSection>

      <DsSection title="Dropdown 下拉菜单" type="card">
        <view class="example-row">
          <DsDropdown v-model="dropdownValue" :options="dropdownOptions" />
          <DsDropdown v-model="dropdownValue" :options="dropdownOptions" direction="up" />
        </view>
      </DsSection>

      <DsSection title="DsCard 卡片" type="card">
        <DsCard title="卡片标题" shadow>
          <text>这是卡片内容，支持自定义头部、内容和底部插槽。</text>
        </DsCard>
        <DsCard title="无边框" bordered :size="'small'">
          <text>小型卡片内容。</text>
        </DsCard>
        <DsCard :size="'large'">
          <text>大型卡片内容，padding更大。</text>
          <template #footer>
            <DsButton text="操作按钮" size="small" />
          </template>
        </DsCard>
      </DsSection>

      <DsSection title="资料卡片组合" type="card">
        <DsCard>
          <view class="example-profile-card">
            <DsAvatar size="large" />
            <view class="example-profile-main">
              <view class="example-profile-title">
                <text>张三</text>
                <DsTag text="VIP" type="primary" size="small" />
              </view>
              <text class="example-profile-subtitle">产品经理</text>
            </view>
            <DsButton text="联系" type="ghost" size="small" />
          </view>
          <DsDescription>
            <DsDescriptionItem label="电话" value="13800138000" />
            <DsDescriptionItem label="邮箱" value="zhangsan@example.com" />
          </DsDescription>
        </DsCard>
      </DsSection>

      <DsSection title="Description 描述列表" type="card">
        <DsDescription title="个人信息">
          <DsDescriptionItem label="姓名" value="张三" />
          <DsDescriptionItem label="年龄" value="28" />
          <DsDescriptionItem label="职位" value="产品经理" />
          <DsDescriptionItem label="部门" value="产品部" />
        </DsDescription>
      </DsSection>

      <DsSection title="DsProgress 进度条" type="card">
        <DsProgress :value="50" label="基础进度" />
        <DsProgress :value="80" type="success" />
        <DsProgress :value="30" type="warning" />
        <DsProgress :value="100" type="danger" showPercent />
        <DsProgress :value="60" :strokeWidth="20" color="#999" />
        <DsProgress :value="45" :strokeWidth="8" type="success" />
      </DsSection>

      <DsSection title="DsStatistic 统计" type="card">
        <view class="example-grid">
          <view class="stat-item">
            <DsStatistic :value="1234" label="用户数" icon="my" />
          </view>
          <view class="stat-item">
            <DsStatistic :value="5678" label="订单数" icon="list" />
          </view>
        </view>
        <view class="example-grid">
          <view class="stat-item">
            <DsStatistic :value="99.9" label="成功率" :precision="1" icon="heart" />
          </view>
          <view class="stat-item">
            <DsStatistic :value="123.45" prefix="¥" label="金额" icon="wallet" />
          </view>
        </view>
        <view class="example-grid">
          <view class="stat-item">
            <DsStatistic :value="0" label="暂无数据" />
          </view>
          <view class="stat-item">
            <DsStatistic :value="1000000" label="总数量" :precision="0" />
          </view>
        </view>
      </DsSection>

      <DsSection title="DsTag 标签" type="card">
        <view class="example-row">
          <DsTag text="默认" />
          <DsTag text="主要" type="primary" />
          <DsTag text="成功" type="success" />
          <DsTag text="警告" type="warning" />
          <DsTag text="危险" type="danger" />
        </view>
        <view class="example-row">
          <DsTag text="小型" size="small" />
          <DsTag text="中型" size="medium" />
          <DsTag text="大型" size="large" />
        </view>
        <view class="example-row">
          <DsTag text="带图标" icon="warning" />
          <DsTag text="可关闭" closable @close="handleTagClose" />
          <DsTag text="禁用" disabled />
        </view>
      </DsSection>

      <DsSection title="SearchBar 搜索框" type="card">
        <DsSearchBar v-model="searchValue" placeholder="请输入搜索内容" />
        <DsSearchBar v-model="searchValue" placeholder="搜索" actionText="搜索" />
        <DsSearchBar v-model="searchValue" placeholder="搜索" actionIcon="qr-code" />
        <DsSearchBar v-model="searchValue" placeholder="搜索" :showCancel="true" />
      </DsSection>

      <DsSection title="DsSlider 滑块" type="card">
        <DsSlider v-model="sliderValue" :min="0" :max="100" unit="%" />
        <DsSlider v-model="sliderValue2" :min="0" :max="100" showLabel label="音量" />
        <DsSlider v-model="sliderValue3" :min="0" :max="200" showRange color="#67c23a" />
        <DsSlider v-model="sliderValue4" :min="0" :max="100" disabled />
      </DsSection>

      <DsSection title="DsFab 悬浮按钮" type="card">
        <view class="fab-demo">
          <DsFab icon="add" />
          <DsFab icon="add" mini />
          <DsFab icon="edit" position="left-bottom" />
        </view>
      </DsSection>

      <DsSection title="NoticeBar 通知栏" type="card">
        <DsNoticeBar text="这是一条通知消息，支持滚动展示较长内容。" />
        <DsNoticeBar text="成功通知" type="success" />
        <DsNoticeBar text="警告通知" type="warning" closable />
        <DsNoticeBar text="危险通知" type="danger" />
        <DsNoticeBar text="信息通知" type="info" icon="information" />
      </DsSection>

      <DsSection title="Image 图片" type="card">
        <view class="example-grid">
          <view class="image-item">
            <DsImage src="/static/logo.png" mode="aspectFill" />
            <text>aspectFill</text>
          </view>
          <view class="image-item">
            <DsImage src="/static/logo.png" mode="aspectFit" />
            <text>aspectFit</text>
          </view>
          <view class="image-item">
            <DsImage src="/static/logo.png" mode="widthFix" />
            <text>widthFix</text>
          </view>
        </view>
      </DsSection>

      <DsSection title="Swiper 轮播" type="card">
        <DsSwiper :items="swiperItems" autoplay pagination />
        <DsSwiper :items="swiperItems" :indicatorDots="true" circular />
        <DsSwiper autoplay :interval="2000">
          <DsSwiperItem>
            <view class="swiper-custom-item">
              <text>自定义内容1</text>
            </view>
          </DsSwiperItem>
          <DsSwiperItem>
            <view class="swiper-custom-item">
              <text>自定义内容2</text>
            </view>
          </DsSwiperItem>
          <DsSwiperItem>
            <view class="swiper-custom-item">
              <text>自定义内容3</text>
            </view>
          </DsSwiperItem>
        </DsSwiper>
      </DsSection>

      <DsSection title="DsCollapse 折叠面板" type="card">
        <DsCollapse v-model="collapseValue">
          <DsCollapseItem title="折叠项1" :index="0">
            <text>折叠内容1</text>
          </DsCollapseItem>
          <DsCollapseItem title="折叠项2" :index="1">
            <text>折叠内容2</text>
          </DsCollapseItem>
          <DsCollapseItem title="折叠项3" :index="2">
            <text>折叠内容3</text>
          </DsCollapseItem>
        </DsCollapse>
        <DsCollapse v-model="collapseValue2" accordion>
          <DsCollapseItem title="手风琴1" :index="0">
            <text>手风琴模式，同时只能展开一个</text>
          </DsCollapseItem>
          <DsCollapseItem title="手风琴2" :index="1">
            <text>手风琴模式，同时只能展开一个</text>
          </DsCollapseItem>
        </DsCollapse>
      </DsSection>

      <DsSection title="DsQrCode 二维码" type="card">
        <view class="example-grid">
          <view class="qr-item">
            <DsQrCode value="https://example.com" :size="150" />
          </view>
          <view class="qr-item">
            <DsQrCode value="https://example.com" :size="200" text="扫码访问" />
          </view>
          <view class="qr-item">
            <DsQrCode value="Hello World" :size="180" level="L" />
          </view>
        </view>
      </DsSection>

      <DsSection title="DsEmpty 空状态" type="card">
        <DsEmpty text="暂无数据" />
        <DsEmpty type="search" text="未找到相关内容" />
        <DsEmpty type="error" text="加载失败" subText="请稍后重试" />
        <DsEmpty type="network" text="网络异常" />
      </DsSection>

      <DsSection title="DsLoading 加载" type="card">
        <view class="example-row">
          <DsLoading />
          <DsLoading type="spinner" text="加载中..." />
          <DsLoading type="dots" text="加载中..." />
        </view>
        <view class="example-row">
          <DsLoading text="加载失败" type="dots" color="#f56c6c" />
          <DsLoading text="自定义颜色" color="#67c23a" />
        </view>
      </DsSection>

      <DsSection title="Result 结果" type="card">
        <DsResult type="success" title="操作成功" message="数据已保存" />
        <DsResult type="fail" title="操作失败" message="请检查后重试" />
        <DsResult type="warning" title="警告提示" message="请确认操作" />
        <DsResult type="info" title="信息提示" message="这是一条信息" />
        <DsResult type="empty" title="暂无数据" />
      </DsSection>

      <DsSection title="弹窗组件" type="card">
        <view class="example-row">
          <DsButton text="对话框" @click="showDialog = true" />
          <DsButton text="抽屉" @click="showDrawer = true" />
        </view>
      </DsSection>

      <DsSection title="Backtop 返回顶部" type="card">
        <DsButton text="滚动后点击右下角返回顶部" @click="scrollToTop" />
      </DsSection>

      <view class="example-space"></view>
    </view>

    <DsDialog v-model:visible="showDialog" title="提示" message="确定要执行此操作吗？" />
    
    <DsDrawer v-model:visible="showDrawer" title="抽屉标题">
      <view class="drawer-content">
        <DsListItem title="选项1" arrow />
        <DsListItem title="选项2" arrow />
        <DsListItem title="选项3" arrow />
      </view>
    </DsDrawer>

    <DsBacktop />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import DsNavBar from '@/components/DsNavBar.vue'
import DsSection from '@/components/DsSection.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsAvatar from '@/components/DsAvatar.vue'
import DsBadge from '@/components/DsBadge.vue'
import DsButton from '@/components/DsButton.vue'
import DsSwitch from '@/components/DsSwitch.vue'
import DsFormItem from '@/components/DsFormItem.vue'
import DsInput from '@/components/DsInput.vue'
import DsInputBar from '@/components/DsInputBar.vue'
import DsNumberInput from '@/components/DsNumberInput.vue'
import DsComboBox from '@/components/DsComboBox.vue'
import DsRadio from '@/components/DsRadio.vue'
import DsCheckbox from '@/components/DsCheckbox.vue'
import DsList from '@/components/DsList.vue'
import DsListItem from '@/components/DsListItem.vue'
import DsGrid from '@/components/DsGrid.vue'
import DsTabs from '@/components/DsTabs.vue'
import DsTabItem from '@/components/DsTabItem.vue'
import DsDropdown from '@/components/DsDropdown.vue'
import DsCard from '@/components/DsCard.vue'
import DsDescription from '@/components/DsDescription.vue'
import DsDescriptionItem from '@/components/DsDescriptionItem.vue'
import DsProgress from '@/components/DsProgress.vue'
import DsStatistic from '@/components/DsStatistic.vue'
import DsTag from '@/components/DsTag.vue'
import DsSearchBar from '@/components/DsSearchBar.vue'
import DsSlider from '@/components/DsSlider.vue'
import DsFab from '@/components/DsFab.vue'
import DsNoticeBar from '@/components/DsNoticeBar.vue'
import DsImage from '@/components/DsImage.vue'
import DsSwiper from '@/components/DsSwiper.vue'
import DsSwiperItem from '@/components/DsSwiperItem.vue'
import DsCollapse from '@/components/DsCollapse.vue'
import DsCollapseItem from '@/components/DsCollapseItem.vue'
import DsQrCode from '@/components/DsQrCode.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsLoading from '@/components/DsLoading.vue'
import DsResult from '@/components/DsResult.vue'
import DsDialog from '@/components/DsDialog.vue'
import DsDrawer from '@/components/DsDrawer.vue'
import DsBacktop from '@/components/DsBacktop.vue'
const switchValue = ref(true)
const inputValue = ref('')
const passwordValue = ref('')
const numberValue = ref(10)
const decimalValue = ref(3.5)
const stepValue = ref(10)
const comboValue = ref('')
const comboValue2 = ref('')
const comboValue3 = ref('2')
const radioValue = ref('')
const radioValue2 = ref('')
const radioValue3 = ref('male')
const checkboxValue = ref([])
const checkboxValue2 = ref([])
const checkboxValue3 = ref([])
const searchValue = ref('')
const sliderValue = ref(50)
const sliderValue2 = ref(70)
const sliderValue3 = ref(100)
const sliderValue4 = ref(30)
const activeTab = ref(0)
const activeTab2 = ref(0)
const activeTab3 = ref(0)
const activeTab4 = ref(0)
const dropdownValue = ref('')
const showDialog = ref(false)
const showDrawer = ref(false)
const collapseValue = ref([0])
const collapseValue2 = ref(0)

const comboOptions = [
  { label: '选项1', value: '1' },
  { label: '选项2', value: '2' },
  { label: '选项3', value: '3' }
]

const radioOptions = [
  { label: '男', value: 'male' },
  { label: '女', value: 'female' }
]

const radioOptionsWithIcon = [
  { label: '微信', value: 'wechat', icon: 'message' },
  { label: '支付宝', value: 'alipay', icon: 'wallet' }
]

const checkboxOptions = [
  { label: '选项1', value: '1' },
  { label: '选项2', value: '2' },
  { label: '选项3', value: '3' }
]

const checkboxOptionsWithIcon = [
  { label: '苹果', value: 'apple', icon: 'store' },
  { label: '香蕉', value: 'banana', icon: 'store' },
  { label: '橙子', value: 'orange', icon: 'store' }
]

const gridItems = [
  { icon: 'home', text: '首页', color: '#5BA3E8' },
  { icon: 'member', text: '个人', color: '#67c23a' },
  { icon: 'store', text: '商城', color: '#e6a23c' },
  { icon: 'message', text: '消息', color: '#f56c6c' }
]

const gridItems2 = [
  { icon: 'order', text: '订单', color: '#5BA3E8' },
  { icon: 'wallet', text: '钱包', color: '#f56c6c' },
  { icon: 'service', text: '客服', color: '#67c23a' }
]

const gridItems3 = [
  { icon: 'setup', text: '设置', color: '#5BA3E8' },
  { icon: 'share', text: '分享', color: '#e6a23c' }
]

const dropdownOptions = [
  { label: '选项A', value: 'A' },
  { label: '选项B', value: 'B' },
  { label: '选项C', value: 'C' }
]

const swiperItems = [
  { image: '/static/logo.png', text: '横幅1' },
  { image: '/static/logo.png', text: '横幅2' },
  { image: '/static/logo.png', text: '横幅3' }
]

const handleTagClose = () => {
  uni.showToast({ title: '标签已关闭', icon: 'none' })
}

const scrollToTop = () => {
  uni.pageScrollTo({ scrollTop: 0, duration: 300 })
}
</script>

<style lang="scss">
@import '@/components/variables.scss';

.examples-page {
  min-height: 100vh;
  background: $bg-secondary;
}

.examples-content {
  padding-bottom: 120rpx;
}

.example-grid {
  display: flex;
  justify-content: space-around;
  
  .example-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: $spacing-xs;
    
    text {
      font-size: $font-size-xs;
      color: $text-secondary;
    }
  }

  .stat-item {
    flex: 1;
    padding: $spacing-sm;
  }

  .qr-item {
    flex: 1;
    display: flex;
    justify-content: center;
  }

  .image-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: $spacing-xs;
    padding: $spacing-sm;

    text {
      font-size: $font-size-xs;
      color: $text-secondary;
    }
  }
}

.example-row {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-sm;
  align-items: center;
}

.example-profile-card {
  display: flex;
  align-items: center;
  gap: $spacing-base;
  margin-bottom: $spacing-base;
}

.example-profile-main {
  flex: 1;
  min-width: 0;
}

.example-profile-title {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  margin-bottom: 8rpx;

  text {
    font-size: $font-size-lg;
    font-weight: 600;
    color: $text-primary;
  }
}

.example-profile-subtitle {
  font-size: $font-size-sm;
  color: $text-secondary;
}

.example-space {
  height: 120rpx;
}

.drawer-content {
  padding: $spacing-base;
}

.fab-demo {
  height: 200rpx;
  position: relative;
}

.swiper-custom-item {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 400rpx;
  background: linear-gradient(135deg, #5BA3E8, #67c23a);
  color: #fff;
  font-size: $font-size-lg;
}
</style>
