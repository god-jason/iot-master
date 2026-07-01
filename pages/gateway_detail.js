// 网关详情页面配置
// 功能：查看网关详细信息，支持编辑、删除操作
return {
  title: '网关详情',
  icon: '/emoji/antenna.svg',
  template: 'detail',
  toolbar: [
    {
      icon: 'edit',
      type: 'button',
      label: '编辑',
      action: {
        type: 'dialog',
        page: 'device_edit',
        params(data) {
          return { id: data.id }
        },
        after_close(result, data, index) {
          this.load()
        }
      }
    },
    {
      icon: 'delete',
      type: 'button',
      label: '删除',
      confirm: '确认删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/device/delete/' + data.id).subscribe(res => {
            this.notification.success('提示', '删除成功')
            this.navigate('/page/device')
          })
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID', type: 'text' },
    { key: 'name', label: '名称', type: 'text' },
    { key: 'description', label: '说明', type: 'text' },
    {
      key: 'group_id',
      key2: 'group_name',
      label: '组织',
      type: 'text',
      action: {
        type: 'page',
        page: 'group_detail',
        params(data) {
          return { id: data.group_id }
        }
      }
    },
    {
      key: 'product_id',
      key2: 'product_name',
      label: '产品',
      type: 'text',
      action: {
        type: 'page',
        page: 'product_detail',
        params(data) {
          return { id: data.product_id }
        }
      }
    },
    { key: 'link_id', label: '连接ID', type: 'text' },
    { key: 'online', label: '在线', type: 'boolean' },
    { key: 'error', label: '错误', type: 'boolean' },
    { key: 'error_string', label: '错误内容', type: 'text' },
    { key: 'location', label: '详细位置', type: 'text' },
    { key: 'longitude', label: '经度', type: 'number' },
    { key: 'latitude', label: '纬度', type: 'number' },
    { key: 'geo_code', label: 'Geo Hash', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
    { key: 'created', label: '创建时间', type: 'datetime' },
    { key: 'updated', label: '更新时间', type: 'datetime' }
  ],
  load_api: 'table/device/detail/:id',
  load_success(data) {
    this.load_product()
  },
  methods: {
    load_product() {
      if (!this.product && this.data.product_id)
        this.request.get('table/product/detail/' + this.data.product_id).subscribe(res => {
          this.product = res.data
          this.add_tabs(res.data)
        })
    },
    add_tabs(data) {
      this.content.tabs = [
        {
          title: '数据',
          icon: '/emoji/activity.svg',
          page: 'device_values',
          params: { id: this.params.id, product_id: data.id }
        },
        {
          title: '日志',
          icon: '/emoji/log.svg',
          page: 'device_log',
          params: { id: this.params.id, product_id: data.id }
        },
        {
          title: '告警',
          icon: '/emoji/alert.svg',
          page: 'alarm',
          params: { device_id: this.params.id, product_id: data.id }
        },
        {
          title: '历史轨迹',
          icon: '/emoji/location.svg',
          page: 'device_track',
          params: { id: this.params.id, product_id: data.id }
        },
        {
          title: '参数配置',
          icon: '/emoji/setting.svg',
          page: 'device_settings',
          params: { id: this.params.id, product_id: data.id }
        },
        {
          title: '设备操作',
          icon: '/emoji/action.svg',
          page: 'device_actions',
          params: { id: this.params.id, product_id: data.id }
        },
        {
          title: '子设备',
          icon: '/emoji/radio.svg',
          page: 'gateway_device',
          params: { gateway_id: this.params.id, product_id: data.id }
        },
        {
          title: '内联设备',
          icon: '/emoji/link.svg',
          page: 'inline',
          params: { gateway_id: this.params.id, product_id: data.id }
        },
        {
          title: '场景管理',
          icon: '/emoji/scene.svg',
          page: 'scene',
          params: { gateway_id: this.params.id, product_id: data.id }
        },
        {
          title: '串口管理',
          icon: '/emoji/serial.svg',
          page: 'serial',
          params: { gateway_id: this.params.id, product_id: data.id }
        },
        {
          title: '定时任务',
          icon: '/emoji/timer.svg',
          page: 'job',
          params: { gateway_id: this.params.id, product_id: data.id }
        },
        {
          title: '数据绑定',
          icon: '/emoji/binding.svg',
          page: 'binding',
          params: { gateway_id: this.params.id, product_id: data.id }
        }
      ]
    }
  },
  tabs: []
}
