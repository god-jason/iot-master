// 报警日志页面配置
return {
  title: '报警日志',
  icon: '/icons/alert.svg',
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
      label: '批量删除',
      icon: 'delete',
      type: 'button',
      confirm: '确认批量删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.table.selects.forEach(id =>
            this.request.get('table/alarm/delete/' + id).subscribe(res => {
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
            this.filter.created = undefined
          }
          this.load()
        }
      }
    }
  ],
  keywords: ['title', 'message'],
  operators: [
    {
      icon: 'delete',
      title: '删除',
      confirm: '确认删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/alarm/delete/' + data.id).subscribe(res => {
            this.notification.success('提示', '删除成功')
            this.load()
          })
        }
      }
    }
  ],
  batch: true,
  fields: [
    { key: 'id', label: '编号', sortable: true, sort: -1, type: 'text' },
    { key: 'title', label: '标题', type: 'text' },
    { key: 'message', label: '消息', type: 'text' },
    { key: 'level', label: '等级', type: 'text' },
    { key: 'created', label: '日期', type: 'date', sortable: true },
    {
      key: 'device_name',
      label: '设备名称',
      type: 'text',
      action: {
        type: 'page',
        page: 'device_detail',
        params(data) {
          return { id: data.device_id }
        }
      }
    }
  ],
  search_api: 'table/alarm/search',
  // 页面挂载时执行
  mount() {
    if (this.params.device_id) this.filter.device_id = this.params.device_id
  }
}
