// 设备日志页面配置
return {
  title: '设备日志',
  icon: '/emoji/log.svg',
  template: 'list',
  toolbar: [
    {
      type: 'button',
      icon: 'reload',
      label: '刷新',
      action: {
        type: 'script',
        script(data, index) {
          this.load()
        }
      }
    },
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
  keywords: ['id', 'name', 'description'],
  operators: [
    {
      icon: 'delete',
      title: '删除',
      confirm: '确认删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/device_log/delete/' + data.id).subscribe(res => {
            this.load()
          })
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID', sortable: true, sort: -1, type: 'text' },
    {
      key: 'device_name',
      label: '设备名称',
      type: 'text',
      action: {
        type: 'page',
        page: 'device_detail',
        params(data) {
          return { id: data.device_id }
        }
      }
    },
    {
      key: 'user_name',
      label: '用户名称',
      type: 'text',
      action: {
        type: 'page',
        page: 'user_detail',
        params(data) {
          return { id: data.user_id }
        }
      }
    },
    { key: 'content', label: '日志', type: 'text' },
    { key: 'created', label: '创建时间', type: 'date', sortable: true }
  ],
  search_api: 'table/device_log/search',
  count_api: 'table/device_log/count',
  // 页面挂载时执行
  mount() {
    this.filter.device_id = this.params.id
  }
}
