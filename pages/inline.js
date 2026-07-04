// 内联设备页面配置
return {
  title: '内联设备',
  icon: '/emoji/inline.svg',
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
      admin: true,
      action: {
        type: 'dialog',
        page: 'inline_create',
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
          this.request.get('device/' + this.params.gateway_id + '/download/inline').subscribe(res => {})
        }
      }
    },
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
        page: 'inline_edit',
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
      admin: true,
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/inline/delete/' + data.id + '/' + data.gateway_id).subscribe(res => {
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
    { key: 'description', label: '说明', type: 'text' },
    {
      key: 'product_name',
      label: '产品名称',
      type: 'text',
      action: {
        type: 'page',
        page: 'product_detail',
        params(data) {
          return { id: data.product_id }
        }
      }
    },
    {
      key: 'gateway_name',
      label: '网关名称',
      type: 'text',
      action: {
        type: 'page',
        page: 'device_detail',
        params(data) {
          return { id: data.gateway_id }
        }
      }
    },
    { key: 'disabled', label: '禁用', type: 'boolean' }
  ],
  search_api: 'table/inline/search',
  count_api: 'table/inline/count',
  // 页面挂载时执行
  mount() {
    if (this.params.gateway_id) this.filter.gateway_id = this.params.gateway_id
    this.get_extend_fields()
  },
  methods: {
    from_gateway() {
      this.request
        .post('device/' + this.params.gateway_id + '/action/database', {
          operator: 'find',
          database: 'inline'
        })
        .subscribe(res => {
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
      this.data.map(s => this.request.get('table/inline/delete/' + s.id + '/' + s.gateway_id).subscribe(() => {}))
      this.data = []
    },
    insert_all() {
      this.data.map(s => (s.gateway_id = this.params.gateway_id))
      this.data.map(s => this.request.post('table/inline/create', s).subscribe(() => {}))
    },
    get_extend_fields() {
      this.request.get('device/extend/fields').subscribe(res => {
        if (res.error) return
        ;(res.data || []).map(f => this.content.fields.push(f))
      })
    }
  }
}
