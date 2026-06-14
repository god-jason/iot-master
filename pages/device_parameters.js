// 设备参数页面配置
return {
  tabs: [],
  children: [],
  setting_content: {
    title: '变量',
    icon: '/icons/device.svg',
    template: 'edit',
    load_api: 'device/:id/values',
    submit_api: 'device/:id/write'
  },
  // 页面挂载时执行
  mount() {
    this.load_model()
    console.log('params:', this.params)
  },
  methods: {
    load_model() {
      this.request.get('product/' + this.params.product_id + '/setting/parameter').subscribe(res => {
        if (res.error) return
        if (res.data.content) this.render_settings(res.data.content)
      })
    },
    render_settings(settings) {
      if (!settings) return
      settings.map(p => {
        this.content.tabs.push({
          title: p.label || p.name,
          params: { id: this.params.id, name: p.name },
          content: Object.assign({}, this.content.setting_content, {
            title: p.label,
            fields: p.fields
          })
        })
      })
    }
  }
}
