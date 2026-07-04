// 用户选择页面配置
return {
  title: '用户选择',
  icon: '/emoji/user.svg',
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
          this.search()
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
    { key: 'username', label: '登录名', type: 'text' },
    { key: 'email', label: '邮箱', type: 'text' },
    { key: 'cellphone', label: '手机号', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' }
  ],
  search_api: 'table/user/search',
  count_api: 'table/user/count'
}
