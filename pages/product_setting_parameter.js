// 编辑产品变量页面配置
return {
  title: '编辑产品变量',
  icon: '/icons/box.svg',
  template: 'edit',
  toolbar: [
    {
      type: 'button',
      label: '导出JSON',
      action: {
        type: 'script',
        script(data, index) {
          this.export_json(this.editor.values, 'product-setting-' + this.params.id)
        }
      }
    },
    {
      type: 'button',
      label: '导入JSON',
      action: {
        type: 'script',
        script(data, index) {
          this.import_json().then(data => (this.editor.values = data))
        }
      }
    }
  ],
  fields: [
    {
      key: 'content',
      label: '【配置变量】',
      span: 24,
      type: 'list',
      children: [
        { key: 'name', label: '名称', type: 'text' },
        { key: 'label', label: '显示', type: 'text' },
        { key: 'hidden', label: '隐藏', type: 'switch' },
        {
          key: 'fields',
          label: '配置项',
          span: 24,
          type: 'table',
          children: [
            { key: 'key', label: '变量', type: 'text' },
            { key: 'label', label: '显示名称', type: 'text' },
            { key: 'placeholder', label: '提示', type: 'text' },
            {
              key: 'type',
              label: '数据类型',
              type: 'select',
              default: 'number',
              options: [
                { label: '开关', value: 'switch' },
                { label: '数值', value: 'number' },
                { label: '文本', value: 'text' },
                { label: '文本框', value: 'textarea' },
                { label: '下拉框', value: 'select' },
                { label: '滑块', value: 'slider' },
                { label: '日期', value: 'date' },
                { label: '时间', value: 'time' }
              ]
            },
            {
              key: 'options',
              label: '选项',
              type: 'table',
              children: [
                { key: 'value', label: '值', type: 'number' },
                { key: 'label', label: '名称', type: 'text' }
              ]
            },
            { key: 'min', label: '最小值', type: 'number' },
            { key: 'max', label: '最大值', type: 'number' },
            { key: 'step', label: '步长', type: 'number' }
          ]
        }
      ]
    }
  ],
  load_api: 'product/:id/setting/parameter',
  submit_api: 'product/:id/setting/parameter',
  submit_success(data) {
    this.navigate('/page/product_detail?id=' + this.params.id)
  }
}
