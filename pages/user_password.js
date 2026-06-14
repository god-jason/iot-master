// 修改密码页面配置
return {
  title: '修改密码',
  icon: '/icons/lock.svg',
  template: 'edit',
  fields: [
    { key: 'password', label: '新密码', type: 'password', required: true }
  ],
  submit_api: 'password/:id'
}