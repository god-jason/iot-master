// 组织详情页面配置
return {
  title: '组织详情',
  icon: '/icons/house.svg',
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
            this.navigate('/page/group')
          })
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID' },
    { key: 'user_id', label: '管理员ID' },
    { key: 'user', label: '管理员' },
    { key: 'name', label: '名称' }
  ],
  load_api: 'table/group/detail/:id',
  tabs: [
    {
      title: '组织设备',
      page: 'device',
      params(params) {
        return { group_id: params.id }
      }
    },
    {
      title: '成员管理',
      page: 'member',
      params(params) {
        return { group_id: params.id }
      }
    }
  ]
}