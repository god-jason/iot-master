// 用户管理页面配置
return {
  title: '用户管理',
  icon: '/icons/user.svg',
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
      label: '创建',
      icon: 'plus',
      type: 'button',
      action: {
        type: 'dialog',
        page: 'user_create',
        after_close(result, data, index) {
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
          this.search()
        }
      }
    }
  ],
  keywords: ['id', 'name'],
  operators: [
    {
      icon: 'edit',
      title: '编辑',
      action: {
        type: 'dialog',
        page: 'user_edit',
        params(data) {
          return { id: data.id }
        },
        after_close(result, data, index) {
          this.load()
        }
      }
    },
    {
      icon: 'lock',
      title: '修改密码',
      action: {
        type: 'dialog',
        page: 'user_password',
        params(data) {
          return { id: data.id }
        }
      }
    },
    {
      icon: 'delete',
      title: '删除',
      confirm: '确认删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/user/delete/' + data.id).subscribe(res => {
            this.load()
          })
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID', sortable: true, type: 'text' },
    { key: 'name', label: '名称', sortable: true, type: 'text' },
    { key: 'email', label: '邮箱', type: 'text' },
    { key: 'cellphone', label: '手机号', type: 'text' },
    { key: 'admin', label: '管理员', type: 'boolean', admin: true },
    { key: 'disabled', label: '禁用', type: 'boolean' }
  ],
  search_api: 'table/user/search'
}
