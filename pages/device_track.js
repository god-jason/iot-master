// 设备轨迹页面配置
return {
  title: '设备轨迹',
  template: 'amap',
  type: 'line',
  height: 600,
  toolbar: [],
  // 页面挂载时执行
  mount() {
    this.device_id = this.params.id
    this.loadTrack()
  },
  methods: {
    loadTrack() {
      this.request.post('table/location/search', {
        filter: { device_id: this.device_id },
        limit: 1000,
        sort: { created: -1 }
      }).subscribe(res => this.render(res.data))
    }
  }
}