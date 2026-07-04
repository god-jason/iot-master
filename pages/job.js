// 定时任务页面配置
return {
  title: '定时任务',
  icon: '/emoji/job.svg',
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
            this.request.get('table/job/delete/' + id + '/' + this.params.gateway_id).subscribe(res => {
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
          this.request.get('device/' + this.params.gateway_id + '/download/job').subscribe(res => {})
        }
      }
    }
  ],
  keywords: ['id', 'name'],
  operators: [
    {
      icon: 'delete',
      title: '删除',
      confirm: '确认删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/job/delete/' + data.id + '/' + data.gateway_id).subscribe(res => {
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
    { key: 'time', label: '时间', type: 'text' },
    { key: 'weekdays', label: '日期', type: 'text' },
    { key: 'action', label: '动作', type: 'text' },
    { key: 'data', label: '参数', type: 'json' },
    { key: 'single', label: '单次', type: 'boolean' },
    { key: 'disabled', label: '禁用', type: 'boolean' },
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
    }
  ],
  search_api: 'table/job/search',
  count_api: 'table/job/count',
  // 页面挂载时执行
  mount() {
    if (this.params.gateway_id) this.filter.gateway_id = this.params.gateway_id
  }
}
