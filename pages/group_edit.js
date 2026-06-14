// 编辑组织页面配置
return {
  title: '编辑组织',
  icon: '/icons/house.svg',
  template: 'edit',
  fields: [
    { key: 'user_id', label: '管理员ID', type: 'text' },
    { key: 'name', label: '名称', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'switch' }
  ],
  load_api: 'table/group/detail/:id',
  submit_api: 'table/group/update/:id'
}
