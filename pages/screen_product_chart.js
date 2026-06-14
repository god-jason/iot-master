// 产品类型图表页面配置
return {
  template: 'chart',
  title: '产品类型',
  icon: '/icons/chart.svg',
  type: 'pie',
  theme: 'dark',
  legend: true,
  tooltip: true,
  style: { margin: '5px' },
  bodyStyle: { color: 'white', padding: 0 },
  mount() {
    this.load_data()
  },
  methods: {
    load_data() {
      this.request
        .post('table/device/group', {
          group_by: 'product_id',
          aggregators: [{ func: 'count', field: 'id', as: 'cnt' }],
          joins: [
            {
              table: 'product',
              local_field: 'product_id',
              foreign_field: 'id',
              field: 'name',
              as: 'product_name'
            }
          ]
        })
        .subscribe(
          res => {
            if (res && res.data && res.data.length > 0) {
              const chartData = res.data.map(item => [item.product_name || '未知产品', item.cnt])
              this.render(chartData)
            } else {
              this.render(this.content.demo)
            }
          },
          err => {
            console.error('加载产品类型数据失败:', err)
            this.render(this.content.demo)
          }
        )
    }
  },
  demo: [
    ['物联小白', 4],
    ['RTU', 2],
    ['其他', 1]
  ]
}
