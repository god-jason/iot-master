// 网关详情页面配置
return {
  title: '设备详情',
  icon: '/icons/device.svg',
  template: 'detail',
  toolbar: [
    {
      icon: 'edit',
      type: 'button',
      label: '编辑',
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
      type: 'button',
      label: '删除',
      confirm: '确认删除？',
      action: {
        type: 'script',
        script(data, index) {
          this.request.get('table/device/delete/' + data.id).subscribe(res => {
            this.navigate('/page/device')
          })
        }
      }
    }
  ],
  fields: [
    { key: 'id', label: 'ID' },
    { key: 'name', label: '名称' },
    { key: 'description', label: '说明' },
    {
      key: 'product_id',
      label: '产品ID',
      action: {
        type: 'page',
        page: 'product_detail',
        params(data) {
          return { id: data.product_id }
        }
      }
    },
    { key: 'online', label: '在线', type: 'boolean' },
    { key: 'error_string', label: '错误' },
    { key: 'disabled', label: '禁用' }
  ],
  load_api: 'table/device/detail/:id',
  tabs: [
    {
      title: '数据',
      page: 'device_values',
      params(params) {
        return { id: params.id }
      }
    },
    {
      title: '日志',
      page: 'device_log',
      params(params) {
        return { id: params.id }
      }
    },
    {
      title: '告警',
      page: 'alarm',
      params(params) {
        return { device_id: params.id }
      }
    },
    {
      title: '历史轨迹',
      page: 'device_track',
      params(params) {
        return { id: params.id }
      }
    },
    {
      title: '参数配置',
      page: 'device_settings',
      params(params) {
        return { id: params.id }
      }
    },
    {
      title: '设备操作',
      page: 'device_actions',
      params(params) {
        return { id: params.id }
      }
    },
    {
      title: '子设备',
      page: 'gateway_device',
      params(params) {
        return { gateway_id: params.id }
      }
    },
    {
      title: '内联设备',
      page: 'inline',
      params(params) {
        return { gateway_id: params.id }
      }
    },
    {
      title: '场景管理',
      page: 'scene',
      params(params) {
        return { gateway_id: params.id }
      }
    },
    {
      title: '串口管理',
      page: 'serial',
      params(params) {
        return { gateway_id: params.id }
      }
    },
    {
      title: '定时任务',
      page: 'job',
      params(params) {
        return { gateway_id: params.id }
      }
    },
    {
      title: '数据绑定',
      page: 'binding',
      params(params) {
        return { gateway_id: params.id }
      }
    },
    {
      title: '数据修改（旧）',
      page: 'device_values_setting',
      params(params) {
        return { id: params.id }
      }
    }
  ]
}