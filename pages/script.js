// 脚本页面配置
return {
  title: '脚本',
  icon: '/icons/code.svg',
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
        page: 'script_create',
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
          this.request.get('device/' + this.params.gateway_id + '/download/script').subscribe(res => {})
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
        page: 'script_edit',
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
          this.request.get('table/script/delete/' + data.id + '/' + data.gateway_id).subscribe(res => {
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
    { key: 'repeat', label: '重复执行', type: 'boolean' },
    { key: 'interval', label: '执行间隔(ms)', type: 'text' },
    { key: 'delay', label: '延迟执行(s)', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
    { key: 'gateway_name', label: '网关名称', type: 'text', action: { type: 'page', page: 'device_detail', params(data) { return { id: data.gateway_id } } } }
  ],
  search_api: 'table/script/search',
  // 页面挂载时执行
  mount() {
    if (this.params.gateway_id) this.filter.gateway_id = this.params.gateway_id
  },
  methods: {
    from_gateway() {
      this.request.post('device/' + this.params.gateway_id + '/action/database', { operator: 'find', database: 'script' }).subscribe(res => {
        this.on_gatway_data(res.data)
      })
    },
    on_gatway_data(ds) {
      if (ds && ds.length) {
        this.delete_all()
        this.data = ds
        setTimeout(() => this.insert_all(), 1000)
      }
    },
    delete_all() {
      this.data.map(s => this.request.get('table/script/delete/' + s.id + '/' + s.gateway_id).subscribe(() => {}))
      this.data = []
    },
    insert_all() {
      this.data.map(s => s.gateway_id = this.params.gateway_id)
      this.data.map(s => this.request.post('table/script/create', s).subscribe(() => {}))
    }
  }
}