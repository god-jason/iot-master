// 组织详情页面配置
// 功能：查看组织详细信息，支持编辑、删除操作
return {
  title: '组织详情',
  icon: '/emoji/building.svg',
  template: 'detail',
  toolbar: [
    {
      icon: 'edit',
      type: 'button',
      label: '编辑',
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
      type: 'button',
      label: '删除',
      confirm: '确认删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/group/delete/' + data.id).subscribe(res => {
            this.notification.success('提示', '删除成功')
            this.navigate('/page/group')
          })
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID', type: 'text' },
    {
      key: 'user_id',
      key2: 'user_name',
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
      key: 'agent_id',
      key2: 'agent_name',
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
    { key: 'name', label: '名称', type: 'text' },
    { key: 'cellphone1', label: '电话1', type: 'text' },
    { key: 'cellphone2', label: '电话2', type: 'text' },
    { key: 'cellphone3', label: '电话3', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
    { key: 'created', label: '创建时间', type: 'datetime' }
  ],
  load_api: 'table/group/detail/:id',
  tabs: [
    {
      title: '组织设备',
      icon: '/emoji/radio.svg',
      page: 'device',
      params(data) {
        return { group_id: data.id }
      }
    },
    {
      title: '成员管理',
      icon: '/emoji/users.svg',
      page: 'member',
      params(data) {
        return { group_id: data.id }
      }
    }
  ]
}
