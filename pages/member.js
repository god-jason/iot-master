// 组织用户页面配置
return {
  title: '组织用户',
  icon: '/icons/people.svg',
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
    {
      label: '绑定',
      icon: 'plus',
      type: 'button',
      action: {
        type: 'dialog',
        page: 'user_choose',
        after_close(result, data, index) {
          if (result) {
            this.request.post('table/member/create', { group_id: this.params.group_id, user_id: result.id }).subscribe(res => {
              this.load()
            })
          }
        }
      }},
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
  operators: [
    {
      icon: 'delete',
      title: '解绑',
      confirm: '确认解绑？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/member/delete/' + data.group_id + '/' + data.user_id).subscribe(res => {
            this.load()
          })
        }
      }
    }
  ],
  fields: [
    {
      key: 'user_name',
      label: '用户名',
      sortable: true,
      type: 'text',
      action: {
        type: 'page',
        page: 'user_detail',
        params(data) {
          return { id: data.user_id }
        }
      }
    },
    { key: 'group_name', label: '组织名称', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
    { key: 'created', label: '创建时间', type: 'date', sortable: true }
  ],
  search_api: 'table/member/search',
  // 页面挂载时执行
  mount() {
    this.filter = { group_id: this.params.group_id }
  }
}