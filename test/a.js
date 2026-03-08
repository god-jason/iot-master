this.loaded_model=true;
this.request.get('iot/product/'+this.data.product_id+'/model')
    .subscribe(res=>{if(res.error)return;
        res.data.properties.map(p=>this.content.tabs.unshift({title:p.name,
            items: p.points.map(p=>{return{key:p.name,label:p.label,suffix:p.unit,span:6,action:this.value_action}}) })) })