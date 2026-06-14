// 产品详情页面配置
return {
  title: '产品详情',
  template: 'detail',
  toolbar: [
    {
      icon: 'edit',
      type: 'button',
      label: '编辑',
      action: {
        type: 'dialog',
        page: 'product_edit',
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
            this.navigate('/page/product')
          })
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID' },
    { key: 'name', label: '名称' },
    { key: 'description', label: '说明' },
    { key: 'type', label: '类型' },
    { key: 'version', label: '版本' },
    { key: 'protocol', label: '协议' },
    { key: 'image', label: '图片', type: 'avatar' },
    { key: 'gateway', label: '网关', type: 'boolean' },
    { key: 'smart', label: '智能', type: 'boolean' },
    { key: 'programmable', label: '可编程', type: 'boolean' },
    { key: 'locatable', label: '支持定位', type: 'boolean' },
    { key: 'disabled', label: '禁用', type: 'boolean' }
  ],
  load_api: 'table/product/detail/:id',
  load_success(data) {
    this.add_tab_settings()
    this.add_tab_actions()
    this.add_tab_parameters()
    this.add_tab_validators()
    this.add_tab_versions()
  },
  tabs: [
    {
      title: '产品设备',
      page: 'device',
      params(params) {
        return { product_id: params.id }
      }
    },
    {
      title: '物模型',
      page: 'product_setting_model',
      params(params) {
        return { id: params.id }
      }
    },
    {
      title: '物模型(旧)',
      page: 'product_model',
      params(params) {
        return { id: params.id }
      }
    }
  ],
  methods: {
    add_tab_settings() {
      if (this.data.configurable) {
        this.content.tabs.push({
          title: '配置参数',
          page: 'product_setting_setting',
          params: { id: this.params.id }
        })
      }
    },
    add_tab_actions() {
      if (this.data.controllable) {
        this.content.tabs.push({
          title: '远程操作',
          page: 'product_setting_action',
          params: { id: this.params.id }
        })
      }
    },
    add_tab_parameters() {
      if (this.data.writable) {
        this.content.tabs.push({
          title: '修改变量',
          page: 'product_setting_parameter',
          params: { id: this.params.id }
        })
      }
    },
    add_tab_validators() {
      if (this.data.validate) {
        this.content.tabs.push({
          title: '数值检查',
          page: 'product_setting_validator',
          params: { id: this.params.id }
        })
      }
    },
    add_tab_versions() {
      if (this.data.ota) {
        this.content.tabs.push({
          title: '固件版本',
          page: 'version',
          params: { product_id: this.params.id }
        })
      }
    }
  }
}
