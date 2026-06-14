// 设备编程页面配置
return {
  tabs: [
    {
      title: 'Lua脚本',
      page: 'script',
      params(data) {
        return { gateway_id: this.params.id }
      }
    },
    {
      title: 'Blockly',
      page: 'blockly',
      params(data) {
        return { gateway_id: this.params.id }
      }
    },
    {
      title: 'ST脚本',
      page: 'st',
      params(data) {
        return { gateway_id: this.params.id }
      }
    }
  ]
}