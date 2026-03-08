import {Component, ElementRef, ViewChild} from '@angular/core';
import {load as loadMap} from '@amap/amap-jsapi-loader';
import {NgIf} from '@angular/common';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzIconDirective} from 'ng-zorro-antd/icon';
import {NzSpinComponent} from 'ng-zorro-antd/spin';
import {SmartToolbarComponent} from '../../lib/smart-toolbar/smart-toolbar.component';
import {TemplateBase} from '../template-base.component';
import {AmapContent} from '../template';

@Component({
  selector: 'app-amap',
  standalone: true,
  imports: [
    NgIf,
    NzButtonComponent,
    NzCardComponent,
    NzIconDirective,
    NzSpinComponent,
    SmartToolbarComponent
  ],
  templateUrl: './amap.component.html',
  styleUrl: './amap.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class AmapComponent extends TemplateBase {
  @ViewChild("toolbar", {static: false}) toolbar!: SmartToolbarComponent;
  @ViewChild("mapContainer", {static: false}) mapContainer!: ElementRef;
  toolbarValue = {}

  map: any //AMap.Map;
  mapHeight = "200px"

  AMap!: any //class


  override build() {
    console.log("[amap] build", this.page)
    super.build()

    let content = this.content as AmapContent;
    if (!content) return


    //初始化高度
    if (typeof this.content?.height == "string") {
      this.mapHeight = this.content.height
    } else if (typeof this.content?.height == "number") {
      this.mapHeight = this.content.height + "px"
    } else {
      this.mapHeight = "200px"
    }

    //setTimeout(()=>this.loadMap(), 1500)
    setTimeout(() => this.loadMap(), 50)
  }

  loadMap() {
    console.log("[amap] load real map", this.page)
    let content = this.content as AmapContent;
    if (!content) return

    //@ts-ignore
    window._AMapSecurityConfig = {
      securityJsCode: content.secret || '55de9923dc16159e4750b7c743117e0d',
    };

    //加载地图，并显示
    loadMap({
      key: content.key || 'eb6a831c04b6dfedda190d6254febb58',
      version: '2.0',
      plugins: ['AMap.Icon', 'AMap.Marker', 'AMap.MarkerCluster', 'AMap.MoveAnimation'],
      AMapUI: {
        version: '1.1',
        plugins: [],
      },
    }).then((AMap) => {
      //this.element.nativeElement
      this.map = new AMap.Map(this.mapContainer.nativeElement, {
        //center: [120.301663, 31.574729],  //设置地图中心点坐标
        resizeEnable: true,
        mapStyle: content.style || 'amap://styles/normal',
        zoom: content.zoom || 12,
      });

      // AMap.plugin('AMap.Geocoder', () => {
      //     this.geocoder = new AMap.Geocoder();
      // });
      // this.geocoder = new AMap.Geocoder({ city: '' });
      // this.marker = new AMap.Marker();

      // this.icon = new AMap.Icon({
      //   image: "https://a.amap.com/jsapi_demos/static/demo-center/icons/poi-marker-default.png",
      //   imageSize: new AMap.Size(25, 30), // 图片大小
      // });

      if (content.city)
        this.map.setCity(content.city)

      this.map.setFitView();
    }).catch((e) => {
      console.log(e);
    });
  }


  override render(data: any) {
    console.log('[amap] render', data)
    let content = this.content as AmapContent;
    if (!content) return

    switch (content.type) {
      case "line":
        let path = data?.map((item: any) => item.position || [item.longitude, item.latitude])
        let polyline = new this.AMap.Polyline({path: path})
        this.map.add(polyline);
        break
      case "point":
        let markers = data?.map((item: any) => {
          let marker = new this.AMap.Marker({
            position: item.position || [item.longitude, item.latitude],
            title: item.name || item.id,
          })
          //响应点击事件
          marker.on("click", console.log)
          return marker
        })

        this.map.add(markers);
        break
      case "cluster":
        // let points = data?.map((item: any) => {
        //
        // })
        // new this.AMap.MarkerCluster(map, points, {
        //   gridSize: 80 // 聚合网格像素大小
        // });
        break
      case "animation":
        //绘制历史轨迹

        break
    }

    this.map.setFitView();
  }
}
