// 通道管理页面配置
return {
  title: '通道管理',
  icon: '/icons/antenna.svg',
  template: 'list',
  toolbar: [
    {
      type: 'button',
      icon: 'reload',
      label: '刷新',
      action: {
        type: 'script',
        script(data, index) {
          this.load()
        }
      }
    },
    {
      label: '创建',
      icon: 'plus',
      type: 'button',
      action: {
        type: 'dialog',
        page: 'socket_create',
        params(data) {
          return { gateway_id: this.params.gateway_id }
        },
        after_close(result, data, index) {
          this.load()
        }
      }
    },
    {
      label: '从网关读取',
      icon: 'sync',
      type: 'button',
      action: {
        type: 'script',
        script(data, index) {
          this.load_device_sockets()
        }
      }
    },
    {
      label: '下载到网关',
      icon: 'download',
      type: 'button',
      confirm: '确认下载到网关？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('device/' + this.params.gateway_id + '/download/socket').subscribe(res => {})
        }
      }},
    { key: 'keyword', type: 'text', placeholder: '请输入关键字' },
    { key: 'range', type: 'daterange', placeholder: ['开始日期', '结束日期'] },
    {
      type: 'button',
      icon: 'search',
      label: '搜索',
      action: {
        type: 'script',
        script(data, index) {
          const v = this.toolbar.value || {}
          this.keyword = v.keyword || ''
          if (v.range && v.range[0]) {
            this.filter.created = { $gte: v.range[0], $lte: v.range[1] }
          } else {
            this.filter.created = undefined
          }
          this.load()
        }
      }
    }
  ],
  keywords: ['id', 'name'],
  operators: [
    {
      icon: 'eye',
      action: {
        type: 'dialog',
        page: 'socket_detail',
        params(data) {
          return { id: data.id, gateway_id: data.gateway_id }
        }
      }
    },
    {
      icon: 'edit',
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
      icon: 'delete',
      title: '删除',
      confirm: '确认删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/socket/delete/' + data.id + '/' + data.gateway_id).subscribe(res => {
            this.load()
          })
        }
      }
    }
  ],
  batch: true,
  fields: [
    {
      key: 'id',
      label: 'ID',
      action: {
        type: 'page',
        page: 'socket_detail',
        params(data) {
          return { id: data.id, gateway_id: data.gateway_id }
        }
      },
      type: 'text'
    },
    { key: 'name', label: '名称', type: 'text', action: { type: 'page', page: 'socket_detail', params(data) { return { id: data.id, gateway_id: data.gateway_id } } } },
    { key: 'adapter', label: '网卡', type: 'text' },
    { key: 'host', label: '主机', type: 'text' },
    { key: 'port', label: '端口', type: 'text' },
    { key: 'protocol', label: '协议', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
    { key: 'gateway_name', label: '网关名称', type: 'text', action: { type: 'page', page: 'device_detail', params(data) { return { id: data.gateway_id } } } },
    { key: 'created', label: '创建时间', type: 'date' }
  ],
  search_api: 'table/socket/search',
  // 页面挂载时执行
  mount() {
    if (this.params.gateway_id) this.filter.gateway_id = this.params.gateway_id
  },
  methods: {
    load_device_sockets() {
      this.request.post('device/' + this.params.gateway_id + '/action/database', { operator: 'find', database: 'socket' }).subscribe(res => {
        this.on_action_sockets(res.data)
      })
    },
    on_action_sockets(sockets) {
      if (sockets && sockets.length) {
        this.delete_sockets()
        this.data = sockets
        setTimeout(() => this.insert_sockets(), 1000)
      }
    },
    delete_sockets() {
      this.data.map(s => this.request.get('table/socket/delete/' + s.id + '/' + s.gateway_id).subscribe(() => {}))
      this.data = []
    },
    insert_sockets() {
      this.data.map(s => s.gateway_id = this.params.gateway_id)
      this.data.map(s => this.request.post('table/socket/create', s).subscribe(() => {}))
    }
  }
}