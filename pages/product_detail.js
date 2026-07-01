// 产品详情页面配置
// 功能：查看产品详细信息，支持编辑、删除操作，管理设备、物模型、配置参数等
return {
  title: '产品详情',
  icon: '/emoji/box.svg',
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
          this.request.get('table/product/delete/' + data.id).subscribe(res => {
            this.notification.success('提示', '删除成功')
            this.navigate('/page/product')
          })
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID', type: 'text' },
    { key: 'name', label: '名称', type: 'text' },
    { key: 'description', label: '说明', type: 'text' },
    { key: 'type', label: '类型', type: 'text' },
    { key: 'version', label: '版本', type: 'text' },
    { key: 'protocol', label: '协议', type: 'text' },
    { key: 'image', label: '图片', type: 'image' },
    { key: 'gateway', label: '网关', type: 'boolean' },
    { key: 'smart', label: '智能', type: 'boolean' },
    { key: 'programmable', label: '可编程', type: 'boolean' },
    { key: 'configurable', label: '可配置', type: 'boolean' },
    { key: 'locatable', label: '支持定位', type: 'boolean' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
    { key: 'created', label: '创建时间', type: 'datetime' }
  ],
  load_api: 'table/product/detail/:id',
  tabs: [
    {
      title: '产品设备',
      icon: '/emoji/device.svg',
      page: 'device',
      params(data) {
        return { product_id: data.id }
      }
    },
    {
      title: '物模型',
      icon: '/emoji/model.svg',
      page: 'product_setting_model',
      params(data) {
        return { id: data.id }
      }
    },
    {
      title: '配置参数',
      icon: '/emoji/setting.svg',
      page: 'product_setting_setting',
      params(data) {
        return { id: data.id }
      }
    },
    {
      title: '设备操作',
      icon: '/emoji/action.svg',
      page: 'product_setting_action',
      params(data) {
        return { id: data.id }
      }
    },
    {
      title: '数值检查',
      icon: '/emoji/check.svg',
      page: 'product_setting_validator',
      params(data) {
        return { id: data.id }
      }
    },
    {
      title: '固件版本',
      icon: '/emoji/version.svg',
      page: 'version',
      params(data) {
        return { product_id: data.id }
      }
    },
    {
      title: '物模型(旧)',
      icon: '/emoji/model.svg',
      page: 'product_model',
      params(data) {
        return { id: data.id }
      }
    }
  ]
}
