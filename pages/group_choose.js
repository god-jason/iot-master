// 选择组织页面配置
return {
  title: '组织',
  icon: '/emoji/house.svg',
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
    { key: 'user_name', label: '管理员', type: 'text' },
    { key: 'agent_name', label: '经销商', type: 'text' },
    { key: 'cellphone1', label: '电话1', type: 'text' },
    { key: 'cellphone2', label: '电话2', type: 'text' },
    { key: 'cellphone3', label: '电话3', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' }
  ],
  search_api: 'table/group/search'
}
