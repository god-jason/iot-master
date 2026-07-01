// 创建产品页面配置
// 功能：创建新产品，支持配置产品属性、协议、功能等
return {
  title: '创建产品',
  icon: '/emoji/create.svg',
  template: 'edit',
  fields: [
    { key: 'id', label: 'ID', type: 'text', required: true, placeholder: '请输入产品ID' },
    { key: 'name', label: '名称', type: 'text', required: true, placeholder: '请输入产品名称' },
    { key: 'description', label: '说明', type: 'text', placeholder: '请输入产品说明' },
    { key: 'type', label: '类型', type: 'text', placeholder: '请输入产品类型' },
    { key: 'version', label: '版本', type: 'text', placeholder: '请输入版本号' },
    { key: 'protocol', label: '协议', type: 'select', options: [], placeholder: '请选择协议' },
    { key: 'gateway', label: '网关', type: 'switch' },
    { key: 'smart', label: '智能', type: 'switch' },
    { key: 'programmable', label: '可编程', type: 'switch' },
    { key: 'configurable', label: '可配置', type: 'switch' },
    { key: 'locatable', label: '支持定位', type: 'switch' },
    { key: 'image', label: '图片', type: 'image', upload: '/api/upload' },
    { key: 'disabled', label: '禁用', type: 'switch' }
  ],
  submit_api: 'table/product/create',
  mount() {
    this.load_protocols()
  },
  methods: {
    load_protocols() {
      this.request.get('protocol/list').subscribe(res => {
        this.content.fields[5].options = res.data.map(d => {
          return { value: d.name, label: d.description }
        })
        this.content.fields[5].options.unshift({ value: '', label: '无' })
      })
    }
  }
}
