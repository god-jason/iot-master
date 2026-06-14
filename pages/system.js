// 系统状态页面配置
return {
  children: [
    {
      span: 24,
      content: {
        title: '线程池',
        template: 'statistic',
        fields: [
          { label: '容量', key: 'cap', span: 6 },
          { label: '空闲', key: 'free', span: 6 },
          { label: '运行', key: 'running', span: 6 },
          { label: '等待', key: 'waiting', span: 6 }
        ],
        load_api: 'dash/pool',
        auto_refresh: 5
      }
    },
    {
      span: 12,
      content: {
        template: 'chart',
        title: 'CPU',
        icon: '/icons/cpu.svg',
        type: 'gauge',
        load_api: 'dash/cpu',
        auto_refresh: 5
      }
    },
    {
      span: 12,
      content: {
        template: 'chart',
        title: '内存',
        type: 'gauge',
        load_api: 'dash/memory',
        auto_refresh: 5,
        gauge: { key: 'percent' }
      }
    },
    {
      content: {
        template: 'list',
        title: '硬盘使用',
        span: 24,
        fields: [
          { key: 'path', label: '路径', type: 'text' },
          { key: 'free', label: '可用', type: 'bytes' },
          { key: 'used', label: '已用', type: 'bytes' },
          { key: 'total', label: '全部', type: 'bytes' },
          { key: 'percent', label: '百分比', type: 'progress' }
        ],
        load_api: 'dash/disk'
      }
    },
    {
      span: 24,
      content: {
        template: 'list',
        title: '网络使用',
        icon: '/icons/net.svg',
        fields: [
          { key: 'index', label: '索引', type: 'text' },
          { key: 'name', label: '名称', type: 'text' },
          { key: 'mtu', label: 'MTU', type: 'text' },
          { key: 'hardware_address', label: '硬件地址', type: 'text' },
          { key: 'address', label: 'IP地址', type: 'tags' },
          { key: 'bytes_sent', label: '发送数据', type: 'bytes' },
          { key: 'bytes_recv', label: '接收数据', type: 'bytes' },
          { key: 'packets_sent', label: '发送数据包', type: 'text' },
          { key: 'packets_recv', label: '接收数据包', type: 'text' }
        ],
        load_api: 'dash/net'
      }
    }
  ]
}