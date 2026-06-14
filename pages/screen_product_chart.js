// 产品类型图表页面配置
return {
  template: 'chart',
  title: '产品类型',
  icon: '/icons/chart.svg',
  type: 'pie',
  theme: 'dark',
  legend: true,
  style: { margin: '5px' },
  bodyStyle: { color: 'white', padding: 0 },
  mount() {
    this.render(this.content.demo)
  },
  demo: [
    ['物联小白', 4],
    ['RTU', 2],
    ['其他', 1]
  ]
}
