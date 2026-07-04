// 用户日志页面配置
return {
  title: '用户日志',
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
  keywords: ['id', 'name', 'action'],
  fields: [
    { key: 'id', label: 'ID', sortable: true, sort: -1, type: 'text' },
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
    { key: 'action', label: '操作', type: 'text' },
    { key: 'client', label: '客户端', type: 'text' },
    { key: 'ip', label: 'IP', type: 'text' },
    { key: 'created', label: '日期', type: 'date' }
  ],
  search_api: 'table/user_log/search',
  count_api: 'table/user_log/count'
}
