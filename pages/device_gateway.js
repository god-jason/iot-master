// 网关设备页面配置
return {
  tabs: [
    {
      title: '子设备',
      page: 'gateway_device',
      params(data) {
        return { gateway_id: this.params.id }
      }
    },
    {
      title: '串口管理',
      page: 'serial',
      params(data) {
        return { gateway_id: this.params.id }
      }
    },
    {
      title: '网络通道',
      page: 'socket',
      params(data) {
        return { gateway_id: this.params.id }
      }
    },
    {
      title: '内联设备',
      page: 'inline',
      params(data) {
        return { gateway_id: this.params.id }
      }
    },
    {
      title: '数据绑定',
      page: 'binding',
      params(data) {
        return { gateway_id: this.params.id }
      }
    },
    {
      title: '连接桥接',
      page: 'bridge',
      params(data) {
        return { gateway_id: this.params.id }
      }
    }
  ]
}
