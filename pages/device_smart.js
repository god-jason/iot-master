// 智能网关页面配置
return {
  tabs: [
    {
      title: '定时任务',
      page: 'job',
      params(data) {
        return { gateway_id: this.params.id }
      }
    },
    {
      title: '场景管理',
      page: 'scene',
      params(data) {
        return { gateway_id: this.params.id }
      }
    }
  ]
}
