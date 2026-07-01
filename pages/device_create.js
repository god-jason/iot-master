// 创建设备页面配置
// 功能：创建新的智能设备，支持选择产品、网关等关联信息
return {
  title: '创建设备',
  icon: '/emoji/create.svg',
  template: 'edit',
  toolbar: [],
  fields: [
    { key: 'id', label: 'ID', type: 'text', placeholder: '默认随机ID' },
    { key: 'name', label: '名称', type: 'text', required: true, placeholder: '请输入设备名称' },
    { key: 'description', label: '说明', type: 'text', placeholder: '请输入设备说明' },
    {
      key: 'product_id',
      label: '产品ID',
      type: 'text',
      placeholder: '请输入产品ID',
      clear: true,
      link_text: '选择产品ID',
      link_action: {
        type: 'dialog',
        page: 'product_choose',
        after_close(result, data, index) {
          this.editor.patchValue({ product_id: result.id })
          this.content.fields[3].tips = result.name
        }
      },
      change_action: {
        type: 'script',
        script(data, index) {
          setTimeout(() => this.load_product(), 100)
        }
      }
    },
    {
      key: 'gateway_id',
      label: '网关ID',
      type: 'text',
      placeholder: '请输入网关ID',
      clear: true,
      link_text: '选择网关ID',
      link_action: {
        type: 'dialog',
        page: 'device_choose',
        after_close(result, data, index) {
          this.editor.patchValue({ gateway_id: result.id })
          this.content.fields[4].tips = result.name
        }
      }
    },
    { key: 'link_id', label: '连接ID', type: 'text', placeholder: '请输入连接ID' },
    { key: 'location', label: '详细位置', type: 'text', placeholder: '请输入详细位置信息' },
    {
      key: 'longitude',
      label: '经度',
      type: 'number',
      placeholder: '点击"选择位置"自动填写',
      min: -180,
      max: 180,
      step: 0.000001,
      link_text: '选择位置',
      link_action: {
        type: 'dialog',
        page: 'map_choose',
        after_close(result, data, index) {
          if (result)
            this.editor.patchValue({
              longitude: result.longitude,
              latitude: result.latitude,
              geo_code: result.geo_code,
              location: result.location
            })
        }
      }
    },
    { key: 'latitude', label: '纬度', type: 'number', placeholder: '点击"选择位置"自动填写', min: -90, max: 90, step: 0.000001 },
    { key: 'geo_code', label: 'Geo Hash', type: 'text', placeholder: '自动生成', hidden: true },
    { key: 'disabled', label: '禁用', type: 'switch' }
  ],
  submit_api: 'table/device/create',
  mount() {
    this.data.product_id = this.params.product_id
    this.data.gateway_id = this.params.gateway_id
    this.get_extend_fields()
  },
  methods: {
    get_extend_fields() {
      this.request.get('device/extend/fields').subscribe(res => {
        if (res.error) return
        ;(res.data || []).map(f => this.content.fields.push(f))
      })
    }
  }
}
