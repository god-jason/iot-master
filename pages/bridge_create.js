// 创建连接桥接页面配置
return {
  title: '创建连接桥接',
  icon: '/icons/bridge.svg',
  template: 'edit',
  fields: [
    { key: 'id', label: 'ID', type: 'text', placeholder: '默认随机ID' },
    { key: 'gateway_id', label: '网关ID', type: 'text' },
    { key: 'name', label: '名称', type: 'text' },
    { key: 'link1', label: '连接1', type: 'text' },
    { key: 'link2', label: '连接2', type: 'text' }
  ],
  submit_api: 'table/bridge/create',
  // 页面挂载时执行
  mount() {
    this.data.gateway_id = this.params.gateway_id
  }
}
