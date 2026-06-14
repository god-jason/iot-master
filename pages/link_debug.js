// 串口调试页面配置
return {
  title: '串口调试',
  icon: '/icons/link.svg',
  template: 'detail',
  toolbar: [
    {
      icon: 'play-circle',
      type: 'button',
      label: '开始调试',
      action: {
        type: 'script',
        script(data, index) {
          this.request
            .post('device/' + this.params.gateway_id + '/action/link', {
              operator: 'debug',
              id: this.params.id
            })
            .subscribe(res => {})
        }
      }
    },
    {
      icon: 'stop',
      type: 'button',
      label: '结束调试',
      action: {
        type: 'script',
        script(data, index) {
          this.request
            .post('device/' + this.params.gateway_id + '/action/link', {
              operator: 'debug_stop',
              id: this.params.id
            })
            .subscribe(res => {})
        }
      }
    },
    {
      key: 'type',
      type: 'select',
      default: 'hex',
      options: [
        { label: '二进制', value: 'hex' },
        { label: '文本', value: 'text' }
      ]
    },
    { key: 'content', type: 'text', placeholder: '请输入内容' },
    {
      icon: 'send',
      type: 'button',
      label: '发送',
      action: {
        type: 'script',
        script(data, index) {
          this.request
            .post('device/' + this.params.gateway_id + '/action/link', {
              operator: 'write',
              id: this.params.id,
              type: this.toolbar.value.type,
              content: this.toolbar.value.content
            })
            .subscribe(res => {})
        }
      }
    }
  ],
  children: [
    {
      span: 12,
      content: {
        title: '读取日志',
        template: 'log',
        mount() {
          this.subscribe('device/' + this.params.gateway_id + '/link/' + this.params.id + '/read', this.on_read)
        },
        methods: {
          on_read(data) {
            this.insert('[' + data.length + '] ' + this.toHex(data) + ' (' + data.toString() + ')')
          },
          toHex(data) {
            return Array.from(data)
              .map(b => b.toString(16).padStart(2, '0'))
              .join(' ')
          }
        }
      }
    },
    {
      span: 12,
      content: {
        title: '写入日志',
        template: 'log',
        mount() {
          this.subscribe('device/' + this.params.gateway_id + '/link/' + this.params.id + '/write', this.on_write)
        },
        methods: {
          on_write(data) {
            this.insert('[' + data.length + '] ' + this.toHex(data) + ' (' + data.toString() + ')')
          },
          toHex(data) {
            return Array.from(data)
              .map(b => b.toString(16).padStart(2, '0'))
              .join(' ')
          }
        }
      }
    }
  ]
}
