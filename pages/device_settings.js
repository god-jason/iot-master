// 设备设置页面配置
return {
  tabs: [],
  children: [],
  setting_content: {
    title: '配置项',
    icon: '/icons/device.svg',
    template: 'edit',
    load_api: 'device/:id/setting/:name',
    submit_api: 'device/:id/setting/:name',
    toolbar: [
      {
        type: 'button',
        label: '读取',
        action: {
          type: 'script',
          script(data, index) {
            this.request.post('device/' + this.params.id + '/action/settings', {
              operator: 'read',
              name: this.params.name
            }).subscribe(res => {
              if (res.error) return
              this.data = res.data
            })
          }
        }
      }
    ]
  },
  // 页面挂载时执行
  mount() {
    this.load_model()
    console.log('params:', this.params)
  },
  methods: {
    load_model() {
      this.request.get('product/' + this.params.product_id + '/setting/setting').subscribe(res => {
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