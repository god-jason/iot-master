// 网关设备页面配置
return {
  tabs: [
    {
      title: '子设备',
      page: 'gateway_device',
      params() { return { gateway_id: this.params.id, product_id: this.params.product_id } }
    },
    {
      title: '串口管理',
      page: 'serial',
      params() { return { gateway_id: this.params.id, product_id: this.params.product_id } }
    },
    {
      title: '网络通道',
      page: 'socket',
      params() { return { gateway_id: this.params.id, product_id: this.params.product_id } }
    },
    {
      title: '内联设备',
      page: 'inline',
      params() { return { gateway_id: this.params.id, product_id: this.params.product_id } }
    },
    {
      title: '数据绑定',
      page: 'binding',
      params() { return { gateway_id: this.params.id, product_id: this.params.product_id } }
    },
    {
      title: '连接桥接',
      page: 'bridge',
      params() { return { gateway_id: this.params.id, product_id: this.params.product_id } }
    }
  ]
}
