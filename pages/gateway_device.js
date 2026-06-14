// 子设备页面配置
return {
  title: '子设备',
  icon: '/icons/device.svg',
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
        page: 'device_create',
        params(data) {
          return {
            product_id: this.params.product_id,
            group_id: this.params.group_id,
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
      action: {
        type: 'dialog',
        page: 'device_import',
        params(data) {
          return { gateway_id: this.params.gateway_id }
        }
      }
    },
    {
      label: '导出',
      icon: 'download',
      type: 'button',
      action: {
        type: 'dialog',
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
      confirm: '确认批量删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.table.selects.forEach(id =>
            this.request.get('table/device/delete/' + id).subscribe(res => {
              this.load()
            })
          )
        }
      }
    },
    {
      key: 'online',
      type: 'select',
      label: '状态',
      options: [{ label: '不过滤' }, { label: '在线', value: 1 }, { label: '离线', value: 0 }],
      change_action: {
        type: 'script',
        script(data, index) {
          setTimeout(() => {
            this.filter.online = this.toolbar.value.online
            this.load()
          }, 100)
        }
      }
    },
    {
      key: 'product_id',
      type: 'select',
      label: '产品',
      options: [{ label: '不过滤' }],
      change_action: {
        type: 'script',
        script(data, index) {
          setTimeout(() => {
            this.filter.product_id = this.toolbar.value.product_id
            this.load()
          }, 100)
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
    },
    {
      label: '下载到网关',
      icon: 'download',
      type: 'button',
      confirm: '确认下载到网关？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('device/' + this.params.gateway_id + '/download/device').subscribe(res => {})
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
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/device/delete/' + data.id).subscribe(res => {
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
    { key: 'group_name', label: '组织名称', type: 'text' },
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
    { key: 'online', label: '在线', type: 'boolean', sortable: true },
    { key: 'error_string', label: '错误', type: 'text' },
    { key: 'location', label: '位置', type: 'text' },
    { key: 'longitude', label: '经度', type: 'text' },
    { key: 'latitude', label: '纬度', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' }
  ],
  search_api: 'table/device/search',
  // 页面挂载时执行
  mount() {
    this.filter.gateway_id = this.params.gateway_id
    this.get_extend_fields()
    if (!this.params.product_id) {
      this.request.post('table/product/search', { limit: 999 }).subscribe(res => {
        if (res.error) return
        this.put_products(res.data)
      })
    }
  },
  methods: {
    get_extend_fields() {
      this.request.get('device/extend/fields').subscribe(res => {
        if (res.error) return
        res.data.map(f => this.content.fields.push(f))
      })
    },
    put_products(products) {
      this.content.toolbar[5].options = [{ label: '不过滤' }].concat(
        products.map(p => {
          return { value: p.id, label: p.name }
        })
      )
    }
  }
}
