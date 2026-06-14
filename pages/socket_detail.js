// 通道详情页面配置
return {
  title: '通道详情',
  icon: '/icons/antenna.svg',
  template: 'detail',
  toolbar: [
    {
      icon: 'edit',
      type: 'button',
      label: '编辑',
      action: {
        type: 'dialog',
        page: 'socket_edit',
        params(data) {
          return { id: data.id, gateway_id: data.gateway_id }
        },
        after_close(result, data, index) {
          this.load()
        }
      }
    },
    {
      icon: 'play-circle',
      type: 'button',
      label: '启动',
      confirm: '确认启动？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.post('device/' + this.params.gateway_id + '/action/link', { operator: 'open', id: this.params.id }).subscribe(res => {})
        }
      }
    },
    {
      icon: 'stop',
      type: 'button',
      label: '停止',
      confirm: '确认停止？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.post('device/' + this.params.gateway_id + '/action/link', { operator: 'close', id: this.params.id }).subscribe(res => {})
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID' },
    { key: 'name', label: '名称' },
    { key: 'adapter', label: '网卡' },
    { key: 'host', label: '主机' },
    { key: 'port', label: '端口' },
    { key: 'protocol', label: '协议' },
    { key: 'disabled', label: '禁用', type: 'boolean' }
  ],
  load_api: 'table/socket/detail/:id/:gateway_id',
  tabs: [
    {
      title: '子设备',
      page: 'device',
      params(params) {
        return { gateway_id: params.gateway_id, link_id: params.id }
      }
    },
    {
      title: '调试',
      page: 'link_debug',
      params(params) {
        return { gateway_id: params.gateway_id, id: params.id }
      }
    }
  ]
}