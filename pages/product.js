// 产品管理页面配置
// 功能：产品列表管理，支持创建、编辑、删除产品，管理设备类型等操作
return {
  title: '产品管理',
  icon: '/emoji/box.svg',
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
        page: 'product_create',
        after_close(result, data, index) {
          this.load()
        }
      }
    },
    { label: '导出', icon: 'download', type: 'button', action: { type: 'page', page: 'product_export' } },
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
            this.filter.created = null
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
        page: 'product_detail',
        params(data) {
          return { id: data.id }
        },
        after_close(result, data, index) {
          this.load()
        }
      }
    },
    {
      icon: 'edit',
      action: {
        type: 'dialog',
        page: 'product_edit',
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
          this.request.get('table/product/delete/' + data.id).subscribe(res => {
            this.notification.success('提示', '删除成功')
            this.load()
          })
        }
      }
    }
  ],
  fields: [
    { key: 'image', label: '图片', type: 'avatar' },
    {
      key: 'id',
      label: 'ID',
      action: {
        type: 'page',
        page: 'product_detail',
        params(data) {
          return { id: data.id }
        }
      },
      sortable: true,
      type: 'text'
    },
    {
      key: 'name',
      label: '名称',
      sortable: true,
      type: 'text',
      action: {
        type: 'page',
        page: 'product_detail',
        params(data) {
          return { id: data.id }
        }
      }
    },
    { key: 'description', label: '说明', type: 'text' },
    { key: 'type', label: '类型', type: 'text', filter: [] },
    { key: 'version', label: '版本', type: 'text' },
    { key: 'protocol', label: '协议', sortable: true, type: 'text', filter: [] },
    { key: 'disabled', label: '禁用', type: 'boolean', filter: [{ text: '全部', value: '' }, { text: '禁用', value: 1 }, { text: '正常', value: 0 }] },
    { key: 'created', label: '日期', type: 'date', sortable: true }
  ],
  search_api: 'table/product/search',
  mount() {
    this.request.get('protocol/list').subscribe(res => {
      this.content.fields[6].filter = res.data?.map(d => ({ text: d.description, value: d.name })) || []
    })

    this.request.post('table/product/search', { limit: 999, fields: ['type'] }).subscribe(res => {
      const types = [...new Set(res.data?.map(p => p.type).filter(Boolean))]
      this.content.fields[4].filter = types.map(t => ({ text: t, value: t }))
    })
  }
}
