// 设备详情页面配置
return {
  title: '设备详情',
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
            this.navigate('/page/device')
          })
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID' },
    { key: 'name', label: '名称' },
    { key: 'description', label: '说明' },
    {
      key: 'product_id',
      label: '产品ID',
      action: {
        type: 'page',
        page: 'product_detail',
        params(data) {
          return { id: data.product_id }
        }
      }
    },
    { key: 'product_name', label: '产品名称' },
    {
      key: 'gateway_id',
      label: '网关ID',
      action: {
        type: 'page',
        page: 'device_detail',
        params(data) {
          return { id: data.gateway_id }
        }
      }
    },
    { key: 'gateway_name', label: '网关名称' },
    { key: 'link_id', label: '连接ID', type: 'text' },
    { key: 'online', label: '在线', type: 'boolean' },
    { key: 'error_string', label: '错误' },
    { key: 'disabled', label: '禁用' }
  ],
  load_api: 'table/device/detail/:id',
  load_success(data) {
    this.load_product()
  },
  methods: {
    load_product() {
      if (!this.product) {
        this.request.get('table/product/detail/' + this.data.product_id).subscribe(res => {
          this.product = res.data
          this.add_tabs(res.data)
        })
      }
    },
    add_tabs(data) {
      if (data.writable) this.add_parameter_tabs()
      if (data.gateway) this.add_gateway_tabs()
      if (data.smart) this.add_smart_tabs()
      if (data.programmable) this.add_program_tabs()
      if (data.configurable) this.add_settings_tabs()
      if (data.locatable) this.add_locate_tabs()
    },
    add_gateway_tabs() {
      this.content.tabs.push({ title: '网关', page: 'device_gateway', params: { id: this.params.id } })
    },
    add_smart_tabs() {
      this.content.tabs.push({ title: '智能', page: 'device_smart', params: { id: this.params.id } })
    },
    add_program_tabs() {
      this.content.tabs.push({ title: '编程', page: 'device_program', params: { id: this.params.id } })
    },
    add_parameter_tabs() {
      this.content.tabs.push({ title: '修改', page: 'device_parameters', params: { id: this.params.id, product_id: this.product.id } })
    },
    add_settings_tabs() {
      this.content.tabs.push({ title: '配置', page: 'device_settings', params: { id: this.params.id, product_id: this.product.id } })
    },
    add_locate_tabs() {
      this.content.tabs.push({ title: '定位', page: 'device_track', params: { id: this.params.id } })
    }
  },
  tabs: [
    {
      title: '数据',
      page: 'device_values',
      params(params) {
        return { id: params.id }
      }
    },
    {
      title: '操作',
      page: 'device_actions',
      params(params) {
        return { id: params.id }
      }
    },
    {
      title: '日志',
      page: 'device_log',
      params(params) {
        return { id: params.id }
      }
    },
    {
      title: '告警',
      page: 'alarm',
      params(params) {
        return { device_id: params.id }
      }
    }
  ]
}