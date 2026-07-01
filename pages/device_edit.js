// 设备编辑页面配置
// 功能：编辑已有设备信息，支持修改产品、网关关联等
return {
  title: '编辑设备',
  icon: '/emoji/edit.svg',
  template: 'edit',
  toolbar: [],
  fields: [
    { key: 'id', label: 'ID', type: 'text', placeholder: '设备ID' },
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
  load_api: 'table/device/detail/:id',
  submit_api: 'table/device/update/:id',
  mount() {
    this.get_extend_fields()
  },
  load_success() {
    this.content.fields[3].tips = this.data.product_name
    this.content.fields[4].tips = this.data.gateway_name
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
