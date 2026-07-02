// 设备管理页面配置
// 功能：设备列表管理，支持创建设备、导入导出、搜索过滤等操作
return {
  title: '设备管理',
  icon: '/emoji/radio.svg',
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
        page: 'device_create',
        params(data) {
          return {
            product_id: this.params.product_id,
            gateway_id: this.params.gateway_id
          }
        },
        after_close(result, data, index) {
          this.load()
        }
      }
    },
    {
      label: '导入',
      icon: 'upload',
      type: 'button',
      admin: true,
      action: {
        type: 'page',
        page: 'device_import',
        params(data) {
          return {
            product_id: this.params.product_id,
            gateway_id: this.params.gateway_id
          }
        }
      }
    },
    {
      label: '导出',
      icon: 'download',
      type: 'button',
      admin: true,
      action: {
        type: 'page',
        page: 'device_export',
        params(data) {
          return { filter: JSON.stringify(this.$event.filter) }
        }
      }
    },
    {
      label: '批量删除',
      icon: 'delete',
      type: 'button',
      admin: true,
      confirm: '确认批量删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.table.selects.forEach(id =>
            this.request.get('table/device/delete/' + id).subscribe(res => {
              this.notification.success('提示', '批量删除成功')
              this.load()
            })
          )
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
            this.filter.created = {
              $gte: v.range[0],
              $lte: v.range[1]
            }
          } else {
            delete this.filter.created
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
        page: 'device_detail',
        params(data) {
          return { id: data.id }
        }
      }
    },
    {
      icon: 'edit',
      action: {
        type: 'dialog',
        page: 'device_edit',
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
      admin: true,
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/device/delete/' + data.id).subscribe(res => {
            this.notification.success('提示', '删除成功')
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
        page: 'device_detail',
        params(data) {
          return { id: data.id }
        }
      },
      sortable: true,
      type: 'text'
    },
    {
      key: 'product_id',
      key2: 'product_name',
      label: '产品',
      action: {
        type: 'page',
        page: 'product_detail',
        params(data) {
          return { id: data.product_id }
        }
      },
      sortable: true,
      type: 'text',
      filter: []
    },
    {
      key: 'name',
      label: '名称',
      sortable: true,
      type: 'text',
      action: {
        type: 'page',
        page: 'device_detail',
        params(data) {
          return { id: data.id }
        }
      }
    },
    { key: 'description', label: '说明', type: 'text' },
    {
      key: 'group_id',
      key2: 'group_name',
      label: '组织',
      type: 'text',
      filter: [],
      action: {
        type: 'page',
        page: 'group_detail',
        params(data) {
          return { id: data.group_id }
        }
      }
    },
    {
      key: 'gateway_id',
      key2: 'gateway_name',
      label: '网关',
      type: 'text',
      filter: [],
      action: {
        type: 'page',
        page: 'device_detail',
        params(data) {
          return { id: data.gateway_id }
        }
      }
    },
    { key: 'online', label: '在线', type: 'boolean', sortable: true, filter: [{ text: '全部', value: '' }, { text: '在线', value: 1 }, { text: '离线', value: 0 }] },
    { key: 'error_string', label: '错误', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean', filter: [{ text: '全部', value: '' }, { text: '禁用', value: 1 }, { text: '正常', value: 0 }] },
    { key: 'created', label: '日期', type: 'date', sortable: true }
  ],
  search_api: 'table/device/search',
  mount() {
    if (this.params.link_id) this.filter.link_id = this.params.link_id
    if (this.params.product_id) this.filter.product_id = this.params.product_id
    if (this.params.gateway_id) this.filter.gateway_id = this.params.gateway_id
    if (!this.params.gateway_id) this.content.toolbar.push(this.content.bind)
    this.get_extend_fields()
    if (!this.params.product_id)
      this.request.post('table/product/search', { limit: 999 }).subscribe(res => {
        if (res.error) return
        this.put_products(res.data || [])
      })
    this.request.post('table/group/search', { limit: 999 }).subscribe(res => {
      if (res.error) return
      this.put_groups(res.data || [])
    })
    this.request.post('table/device/search', { limit: 999, filter: { gateway: 1 } }).subscribe(res => {
      if (res.error) return
      this.put_gateways(res.data || [])
    })
  },
  methods: {
    get_extend_fields() {
      this.request.get('device/extend/fields').subscribe(res => {
        if (res.error) return
        ;(res.data || []).map(f => this.content.columns.push(f))
      })
    },
    put_products(products) {
      const list = Array.isArray(products) ? products : []
      this.content.fields[1].filter = list.map(p => {
        return { value: p.id, text: p.name }
      })
    },
    put_groups(groups) {
      const list = Array.isArray(groups) ? groups : []
      this.content.fields[4].filter = list.map(g => {
        return { value: g.id, text: g.name }
      })
    },
    put_gateways(gateways) {
      const list = Array.isArray(gateways) ? gateways : []
      this.content.fields[5].filter = list.map(g => {
        return { value: g.id, text: g.name }
      })
    }
  },
  bind: {
    type: 'button',
    icon: 'plus',
    label: '添加',
    not_admin: true,
    action: {
      type: 'dialog',
      page: 'device_bind',
      after_close(result, data, index) {
        this.load()
      }
    }
  }
}
