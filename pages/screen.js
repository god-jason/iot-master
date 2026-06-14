// 地图大屏页面配置
return {
  template: 'amap',
  height: '100vh',
  full: true,
  satellite: false,
  mapStyle: 'amap://styles/darkblue',
  zoom: 5,
  ready() {
    this.load_data()
  },
  methods: {
    load_devices() {
      this.request.post('table/device/search', { limit: 999999, fields: ['id', 'longitude', 'latitude', 'online', 'name'] }).subscribe(res => {
        this.render_devices(res.data)
      })
    },
    render_devices(devices) {
      this.addClusters(devices)
    }
  },
  overlay: {
    content: {
      template: 'blank',
      children: [
        {
          page: 'screen_title'
        },
        {
          span: 6,
          content: {
            template: 'blank',
            children: [
              {
                page: 'screen_product_chart'
              }
            ]
          }
        },
        {
          span: 12,
          content: {
            template: 'statistic',
            style: { margin: '5px' },
            style2: { color: 'white', background: 'transparent' },
            bodyStyle: { color: 'white', background: 'transparent' },
            fields: [
              { label: '总数', key: 'total' },
              { label: '在线', key: 'online' },
              { label: '故障', key: 'error' }
            ],
            mount() {
              this.data = this.content.demo
            },
            demo: { total: 10, online: 9, error: 0 }
          }
        },
        {
          span: 6,
          content: {
            template: 'blank',
            children: [
              {
                content: {
                  title: '使用统计',
                  icon: '/icons/chart.svg',
                  template: 'chart',
                  style: { margin: '5px' },
                  type: 'bar',
                  theme: 'dark',
                  bodyStyle: { color: 'white', padding: 0 },
                  mount() {
                    this.render(this.content.demo)
                  },
                  demo: [
                    ['一月', 2],
                    ['二月', 5],
                    ['三月', 5],
                    ['四月', 7]
                  ]
                }
              },
              {
                content: {
                  title: '报警日志',
                  template: 'list',
                  style: { margin: '5px' },
                  bodyStyle: { color: 'white', 'background-color': 'black', padding: 0 }
                }
              }
            ]
          }
        }
      ]
    }
  }
}
