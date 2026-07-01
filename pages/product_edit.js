// 产品编辑页面配置
// 功能：编辑产品信息，包括名称、类型、协议等
return {
  title: '编辑产品',
  icon: '/emoji/edit.svg',
  template: 'edit',
  fields: [
    { key: 'name', label: '名称', type: 'text', required: true, placeholder: '请输入产品名称' },
    { key: 'description', label: '说明', type: 'text', placeholder: '请输入产品说明' },
    { key: 'type', label: '类型', type: 'text', placeholder: '请输入产品类型' },
    { key: 'version', label: '版本', type: 'text', placeholder: '请输入版本号' },
    { key: 'protocol', label: '协议', type: 'select', options: [], placeholder: '请选择协议' },
    { key: 'gateway', label: '网关', type: 'switch' },
    { key: 'smart', label: '智能', type: 'switch' },
    { key: 'controllable', label: '远程控制', type: 'switch' },
    { key: 'writable', label: '修改变量', type: 'switch' },
    { key: 'programmable', label: '可编程', type: 'switch' },
    { key: 'configurable', label: '可配置', type: 'switch' },
    { key: 'ota', label: '远程升级', type: 'switch' },
    { key: 'locatable', label: '支持定位', type: 'switch' },
    { key: 'image', label: '图片', type: 'image', upload: '/api/upload' },
    { key: 'disabled', label: '禁用', type: 'switch' }
  ],
  load_api: 'table/product/detail/:id',
  submit_api: 'table/product/update/:id',
  mount() {
    this.load_protocols()
  },
  methods: {
    load_protocols() {
      this.request.get('protocol/list').subscribe(res => {
        this.content.fields[4].options = res.data.map(d => ({ value: d.name, label: d.description }))
        this.content.fields[4].options.unshift({ value: '', label: '无' })
      })
    }
  }
}
