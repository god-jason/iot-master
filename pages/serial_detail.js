// 串口详情页面配置
return {
  title: '串口详情',
  icon: '/icons/serial.svg',
  template: 'detail',
  toolbar: [
    {
      icon: 'edit',
      type: 'button',
      label: '编辑',
      action: {
        type: 'dialog',
        page: 'serial_edit',
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
    { key: 'gateway_id', label: '网关ID' },
    { key: 'name', label: '名称' },
    { key: 'port', label: '序号' },
    { key: 'baud_rate', label: '波特率' },
    { key: 'data_bits', label: '字长' },
    { key: 'stop_bits', label: '停止位' },
    { key: 'parity', label: '检验位' },
    { key: 'rs485_gpio', label: '485GPIO' },
    { key: 'protocol', label: '协议' },
    { key: 'protocol_options', label: '协议参数', type: 'json' },
    { key: 'disabled', label: '禁用', type: 'boolean' }
  ],
  load_api: 'table/serial/detail/:id/:gateway_id',
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
    },
    {
      title: '透传',
      page: 'serial_transport',
      params(params) {
        return { gateway_id: params.gateway_id, id: params.id }
      }
    }
  ]
}