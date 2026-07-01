// 协议库页面配置
return {
  title: '协议库',
  icon: '/emoji/net.svg',
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
  keywords: ['name', 'description', 'author'],
  fields: [
    { key: 'name', label: '名称', type: 'text' },
    { key: 'description', label: '说明', type: 'text' },
    { key: 'version', label: '版本', type: 'text' },
    { key: 'author', label: '作者', type: 'text' },
    { key: 'copyright', label: '版权', type: 'text' }
  ],
  load_api: 'protocol/list'
}
