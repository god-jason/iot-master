// 选择设备页面配置
return {
  title: '设备',
  icon: '/icons/device.svg',
  template: 'list',
  toolbar: [
    { key: 'keyword', type: 'text', placeholder: '请输入关键字' },
    { key: 'range', type: 'daterange', placeholder: ['开始日期', '结束日期'] },
    {
      type: 'button',
      icon: 'search',
      label: '搜索',
      action: {
        type: 'script',
        script(data, index) {
          const v = this.toolbar.value || {}
          this.keyword = v.keyword || ''
          if (v.range && v.range[0]) {
            this.filter.created = { $gte: v.range[0], $lte: v.range[1] }
          } else {
            this.filter.created = undefined
          }
          this.load()
        }
      }
    }
  ],
  keywords: ['id', 'name'],
  operators: [
    {
      icon: 'check',
      label: '选择',
      action: {
        type: 'script',
        script(data, index) {
          this.modelRef.close(data)
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID', sortable: true, type: 'text' },
    { key: 'name', label: '名称', sortable: true, type: 'text' },
    { key: 'description', label: '说明', type: 'text' },
    { key: 'product_name', label: '产品名称', type: 'text' },
    { key: 'group_name', label: '组织名称', type: 'text' },
    { key: 'gateway_name', label: '网关名称', type: 'text' },
    { key: 'online', label: '在线', type: 'boolean', sortable: true },
    { key: 'error_string', label: '错误', type: 'text' },
    { key: 'location', label: '位置', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
    { key: 'created', label: '日期', type: 'date', sortable: true }
  ],
  search_api: 'table/device/search'
}