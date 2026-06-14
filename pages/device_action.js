// 执行动作页面配置
return {
  title: '执行动作',
  icon: '/icons/device.svg',
  template: 'edit',
  fields: [],
  // 页面挂载时执行
  mount() {
    if (this.params.title) this.content.title = this.params.title
    this.content.fields = this.params.parameters || []
    this.content.fields.forEach(f => {
      if (f.data_api) {
        this.request.get(f.data_api).subscribe(res => {
          if (res.error) return
          f.options = res.data
        })
      }
    })
  },
  submit_api: 'device/:id/action/:action'
}