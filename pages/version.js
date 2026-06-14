// 版本页面配置
return {
  title: '版本',
  icon: '/icons/version.svg',
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
        page: 'version_create',
        params(data) {
          return { product_id: this.params.product_id }
        },
        after_close(result, data, index) {
          this.load()
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
        page: 'version_edit',
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
          this.request.get('table/version/delete/' + data.id).subscribe(res => {
            this.load()
          })
        }
      }
    }
  ],
  batch: true,
  fields: [
    { key: 'id', label: 'ID', sortable: true, sort: -1, type: 'text' },
    { key: 'name', label: '名称', sortable: true, type: 'text' },
    { key: 'description', label: '说明', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
    { key: 'created', label: '日期', type: 'date', sortable: true },
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
    }
  ],
  search_api: 'table/version/search',
  // 页面挂载时执行
  mount() {
    if (this.params.product_id) this.filter.product_id = this.params.product_id
  }
}