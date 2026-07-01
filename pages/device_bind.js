// 添加设备页面配置
return {
  title: '添加设备',
  icon: '/emoji/device.svg',
  template: 'edit',
  fields: [{ key: 'id', label: 'ID', type: 'text', placeholder: '请输入设备ID' }],
  submit(data) {
    this.request.get('device/' + data.id + '/bind').subscribe(res => {
      this.modelRef.close()
    })
  }
}
