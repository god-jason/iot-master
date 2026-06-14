// 组织管理页面配置
return {
  title: '组织',
  icon: '/icons/house.svg',
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
        page: 'group_create',
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
          this.load()
        }
      }
    }
  ],
  keywords: ['id', 'name'],
  operators: [
    {
      icon: 'eye',
      action: {
        type: 'dialog',
        page: 'group_detail',
        params(data) {
          return { id: data.id }
        },
        after_close(result, data, index) {
          this.load()
        }
      }
    },
    {
      icon: 'edit',
      action: {
        type: 'dialog',
        page: 'group_edit',
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
      title: '删除',
      confirm: '确认删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/group/delete/' + data.id).subscribe(res => {
            this.load()
          })
        }
      }
    }
  ],
  fields: [
    {
      key: 'id',
      label: 'ID',
      action: {
        type: 'page',
        page: 'group_detail',
        params(data) {
          return { id: data.id }
        }
      },
      sortable: true,
      type: 'text'
    },
    { key: 'name', label: '名称', sortable: true, type: 'text', action: { type: 'page', page: 'group_detail', params(data) { return { id: data.id } } } },
    {
      key: 'user_name',
      label: '管理员',
      type: 'text',
      action: {
        type: 'page',
        page: 'user_detail',
        params(data) {
          return { id: data.user_id }
        }
      }
    },
    {
      key: 'agent_name',
      label: '经销商',
      type: 'text',
      action: {
        type: 'page',
        page: 'user_detail',
        params(data) {
          return { id: data.agent_id }
        }
      }
    },
    { key: 'cellphone1', label: '电话1', type: 'text' },
    { key: 'cellphone2', label: '电话2', type: 'text' },
    { key: 'cellphone3', label: '电话3', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
    { key: 'created', label: '创建时间', type: 'date', sortable: true }
  ],
  search_api: 'table/group/search'
}