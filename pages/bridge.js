// 连接桥接页面配置
return {
  title: '连接桥接',
  icon: '/icons/bridge.svg',
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
        page: 'bridge_create',
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
          this.from_gateway()
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
          this.request.get('device/' + this.params.gateway_id + '/download/bridge').subscribe(res => {})
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
      icon: 'edit',
      action: {
        type: 'dialog',
        page: 'bridge_edit',
        params(data) {
          return { id: data.id }
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
          this.request.get('table/bridge/delete/' + data.id + '/' + data.gateway_id).subscribe(res => {
            this.load()
          })
        }
      }
    }
  ],
  batch: true,
  fields: [
    { key: 'id', label: 'ID', type: 'text' },
    { key: 'name', label: '名称', type: 'text' },
    { key: 'link1', label: '连接1', type: 'text' },
    { key: 'link2', label: '连接2', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
    { key: 'link1_name', label: '连接1名称', type: 'text' },
    { key: 'link2_name', label: '连接2名称', type: 'text' }
  ],
  search_api: 'table/bridge/search',
  // 页面挂载时执行
  mount() {
    if (this.params.gateway_id) this.filter.gateway_id = this.params.gateway_id
  },
  methods: {
    from_gateway() {
      this.request.post('device/' + this.params.gateway_id + '/action/database', { operator: 'find', database: 'bridge' }).subscribe(res => {
        this.on_gateway_data(res.data)
      })
    },
    on_gateway_data(ds) {
      if (ds && ds.length) {
        this.delete_all()
        this.data = ds
        setTimeout(() => this.insert_all(), 1000)
      }
    },
    delete_all() {
      this.data.map(s => this.request.get('table/bridge/delete/' + s.id + '/' + s.gateway_id).subscribe(() => {}))
      this.data = []
    },
    insert_all() {
      this.data.map(s => s.gateway_id = this.params.gateway_id)
      this.data.map(s => this.request.post('table/bridge/create', s).subscribe(() => {}))
    }
  }
}