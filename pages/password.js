// 修改密码页面配置
return {
  title: '修改密码',
  icon: '/icons/lock.svg',
  template: 'edit',
  fields: [
    { key: 'old', label: '旧密码', type: 'password', required: true },
    { key: 'new', label: '新密码', type: 'password', required: true }
  ],
  submit_api: 'password',
  submit_success(data) {
    this.navigate('/login')
  }
}