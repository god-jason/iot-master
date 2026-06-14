import { Component, ElementRef, ViewChild, OnDestroy } from '@angular/core';
import { load as loadMap } from '@amap/amap-jsapi-loader';
import convert from 'wzlcoordconvert';
import { encodeBase32 } from 'geohashing';

import { NzButtonComponent } from 'ng-zorro-antd/button';
import { NzCardComponent } from 'ng-zorro-antd/card';
import { NzIconDirective } from 'ng-zorro-antd/icon';
import { NzSpinComponent } from 'ng-zorro-antd/spin';
import { SmartToolbarComponent } from '../../lib/smart-toolbar/smart-toolbar.component';
import { TemplateBase } from '../template-base.component';
import { AmapContent } from '../template';
import { CommonModule } from '@angular/common';
import { isFunction } from 'rxjs/internal/util/isFunction';

@Component({
  selector: 'app-amap',
  standalone: true,
  imports: [
    CommonModule,
    NzButtonComponent,
    NzCardComponent,
    NzIconDirective,
    NzSpinComponent,
    SmartToolbarComponent
  ],
  templateUrl: './amap.component.html',
  styleUrl: './amap.component.scss',
})
export class AmapComponent extends TemplateBase implements OnDestroy {
  @ViewChild("toolbar", { static: false }) toolbar!: SmartToolbarComponent;
  @ViewChild("mapContainer", { static: false }) mapContainer!: ElementRef;
  toolbarValue = {}

  map: any //AMap.Map 实例
  mapHeight = "200px"

  AMap!: any //AMap类引用

  polygonEditor: any //多边形编辑器

  geocoder: any //地理编码服务

  //存储已添加的地图覆盖物，方便后续管理
  overlays: any[] = []

  /**
   * 组件构建时初始化地图
   */
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

    setTimeout(() => this.loadMap(), 50)
  }

  /**
   * 加载高德地图
   */
  loadMap() {
    console.log("[amap] load real map", this.page)
    let content = this.content as AmapContent;
    if (!content) return

    //@ts-ignore
    window._AMapSecurityConfig = {
      securityJsCode: content.secret || '55de9923dc16159e4750b7c743117e0d',
    };

    loadMap({
      key: content.key || 'eb6a831c04b6dfedda190d6254febb58',
      version: '2.0',
      plugins: ['AMap.Icon', "AMap.Circle", 'AMap.CircleMarker', 'AMap.BezierCurve', 'AMap.Marker', 'AMap.MarkerCluster', 'AMap.MoveAnimation', 'AMap.Polygon', 'AMap.PolygonEditor', 'AMap.Geocoder', 'AMap.Geolocation'],
      AMapUI: {
        version: '1.1',
        plugins: [],
      },
    }).then((AMap) => {
      this.AMap = AMap;

      this.map = new AMap.Map(this.mapContainer.nativeElement, {
        resizeEnable: true,
        mapStyle: content.mapStyle || 'amap://styles/normal',
        zoom: content.zoom || 12,
      });

      //初始化地理编码服务
      this.geocoder = new AMap.Geocoder({ city: content.city || '' });

      //添加卫星图层
      if (content.satellite) {
        this.map.add(new AMap.TileLayer.Satellite())
      }

      //设置城市定位
      if (content.city)
        this.map.setCity(content.city)

      this.map.setFitView();

      //渲染初始数据
      if (this.data)
        this.render(this.data)

      //初始化完成，调用ready回调
      if (typeof content.ready == "string")
        content.ready = new Function(content.ready)
      if (isFunction(content.ready))
        content.ready.call(this)

      //添加地图点击事件监听
      this.map.on('click', (e: any) => {
        if (typeof content.click == "string")
          content.click = new Function(content.click)
        if (isFunction(content.click))
          content.click.call(this, e)
      })
    }).catch((e) => {
      console.log(e);
    });
  }

  /**
   * 根据配置类型渲染数据
   * @param data 待渲染的数据
   */
  override render(data: any) {
    console.log('[amap] render', data)
    let content = this.content as AmapContent;
    if (!content) return

    if (!this.AMap) {
      this.data = data;
      return;
    }

    switch (content.type) {
      case "line":
        this.addPolyline(data);
        break
      case "polygon":
        if (data && data.points) {
          if (typeof data.points == "string")
            data.points = JSON.parse(data.points)
          this.addPolygon(data.points, data)
        } else {
          this.startPolygonEdit()
        }
        break;
      case "polygons":
        if (data && data.length) {
          data.forEach((d: any) => {
            if (typeof d.points == "string")
              d.points = JSON.parse(d.points)
            this.addPolygon(d.points, d)
          })
        } else {
          this.startPolygonEdit()
        }
        break;
      case "point":
        this.addMarkers(data);
        break
      case "cluster":
        this.addMarkersCluster(data);
        break
      case "animation":
        break
    }

    this.fitBounds();
  }

  // ==================== 地图元素操作接口 ====================

  /**
   * 添加折线（轨迹）
   * @param data 轨迹数据，每个元素需包含 longitude(经度), latitude(纬度) - WGS84坐标
   * @param options 可选配置：strokeColor(颜色), strokeWeight(宽度), radius(圆点半径)
   */
  addPolyline(data: any[], options: any = {}) {
    if (!this.map || !data || data.length < 2) return;

    //圆点标注具体位置（WGS84转GCJ02）
    let circles = data.map((item: any, index: number) => {
      let [lng, lat] = this.wgs84ToGcj02(item.longitude, item.latitude);
      let marker = new this.AMap.Circle({
        center: [lng, lat],
        radius: options.radius || 4,
        strokeColor: index == 0 ? "#FF0000" : "#fc1313",
        strokeWeight: index == 0 ? 10 : 0,
        title: item.created,
      })
      marker.on("click", console.log)
      return marker;
    })
    this.map.add(circles);
    this.overlays.push(...circles);

    //绘制路径（WGS84转GCJ02）
    let path = data.map((item: any) => {
      let [lng, lat] = this.wgs84ToGcj02(item.longitude, item.latitude);
      return [lng, lat];
    })
    let polyline = new this.AMap.Polyline({
      path: path,
      strokeColor: options.strokeColor || "#2b8cbe",
      strokeWeight: options.strokeWeight || 4
    })
    this.map.add(polyline);
    this.overlays.push(polyline);
  }

  /**
   * 添加多边形（围栏）
   * @param path 多边形顶点数组，格式：[[lng1, lat1], [lng2, lat2], ...] - WGS84坐标
   * @param extra 附加数据（如id等）
   * @param options 可选配置：strokeColor, fillColor, strokeWeight, editable(是否可编辑)
   */
  addPolygon(path: number[][], extra?: any, options: any = {}) {
    if (!this.map || !path || path.length < 3) {
      console.log("地图未准备好或路径点不足")
      return
    }

    //WGS84转GCJ02
    let gcj02Path = path.map((point: number[]) => {
      return this.wgs84ToGcj02(point[0], point[1]);
    })

    let polygon = new this.AMap.Polygon({
      ...options,
      path: gcj02Path,
      strokeColor: options.strokeColor || '#FF6600',
      fillColor: options.fillColor || 'rgba(255, 102, 0, 0.3)',
      strokeWeight: options.strokeWeight || 2,
      editable: options.editable || false
    })
    this.map.add(polygon)
    this.overlays.push(polygon)

    //附加信息
    polygon.setExtData(extra)

    polygon.on('dblclick', () => {
      if (!this.polygonEditor)
        this.initPolygonEditor()

      this.polygonEditor.close();
      this.polygonEditor.addAdsorbPolygons([polygon]);
      this.polygonEditor.setTarget(polygon);
      this.polygonEditor.open();
    })
  }

  /**
   * 添加多个标注点
   * @param data 点数据数组，每个元素需包含 longitude, latitude，可选 name/id 作为标题 - WGS84坐标
   * @param options 可选配置：icon(图标), onClick(点击回调)
   */
  addMarkers(data: any[], options: any = {}) {
    if (!this.map || !data || data.length === 0) return;

    let markers = data.map((item: any) => {
      let [lng, lat] = this.wgs84ToGcj02(item.longitude, item.latitude);
      let marker = new this.AMap.Marker({
        position: [lng, lat],
        title: item.name || item.id,
        icon: options.icon || undefined,
      })
      marker.on("click", options.onClick || console.log)
      return marker
    })

    this.map.add(markers);
    this.overlays.push(...markers);
  }

  /**
   * 添加聚合点
   * @param data 点数据数组，每个元素需包含 longitude, latitude - WGS84坐标
   * @param options 可选配置：gridSize(聚合网格大小)
   */
  addMarkersCluster(data: any[], options: any = {}) {
    if (!this.map || !data || data.length === 0) return;

    let points = data.map((item: any) => {
      let [lng, lat] = this.wgs84ToGcj02(item.longitude, item.latitude);
      return {
        weight: item.weight || 1,
        lnglat: [lng, lat],
      }
    })

    let cluster = new this.AMap.MarkerCluster(this.map, points, {
      ...options,
      gridSize: options.gridSize || 80
    });

    this.overlays.push(cluster);
  }

  /**
   * 添加单个标注点
   * @param longitude 经度 - WGS84坐标
   * @param latitude 纬度 - WGS84坐标
   * @param title 标题
   * @param options 可选配置：icon, onClick
   */
  addMarker(longitude: number, latitude: number, title?: string, options: any = {}) {
    if (!this.map) return;

    //WGS84转GCJ02
    let [lng, lat] = this.wgs84ToGcj02(longitude, latitude);

    let marker = new this.AMap.Marker({
      ...options,
      title: title || '',
      position: [lng, lat],
      anchor: 'bottom-center',
    })

    if (options.onClick) {
      marker.on("click", options.onClick)
    }

    this.map.add(marker);
    this.overlays.push(marker);
    return marker;
  }

  /**
   * 添加圆形标注点（CircleMarker）
   * @param longitude 经度 - WGS84坐标
   * @param latitude 纬度 - WGS84坐标
   * @param radius 半径
   * @param options 可选配置：strokeColor, fillColor, strokeWeight
   */
  addCircleMarker(longitude: number, latitude: number, radius: number, options: any = {}) {
    if (!this.map) return;

    //WGS84转GCJ02
    let [lng, lat] = this.wgs84ToGcj02(longitude, latitude);

    let circle = new this.AMap.CircleMarker({
      ...options,
      center: [lng, lat],
      radius: radius,
      strokeColor: options.strokeColor || '#FF0000',
      fillColor: options.fillColor || 'rgba(255,0,0,0.3)',
      strokeWeight: options.strokeWeight || 2,
    })

    this.map.add(circle);
    this.overlays.push(circle);
    return circle;
  }

  /**
   * 添加圆形
   * @param longitude 圆心经度 - WGS84坐标
   * @param latitude 圆心纬度 - WGS84坐标
   * @param radius 半径(米)
   * @param options 可选配置：strokeColor, fillColor, strokeWeight
   */
  addCircle(longitude: number, latitude: number, radius: number, options: any = {}) {
    if (!this.map) return;

    //WGS84转GCJ02
    let [lng, lat] = this.wgs84ToGcj02(longitude, latitude);

    let circle = new this.AMap.Circle({
      ...options,
      center: [lng, lat],
      radius: radius,
      strokeColor: options.strokeColor || '#FF0000',
      fillColor: options.fillColor || 'rgba(255,0,0,0.3)',
      strokeWeight: options.strokeWeight || 2,
    })

    this.map.add(circle);
    this.overlays.push(circle);
    return circle;
  }

  /**
   * 添加路径动画
   * @param data 轨迹数据，每个元素需包含 longitude(经度), latitude(纬度) - WGS84坐标
   * @param options 可选配置
   */
  animatePath(data: any[], options: any = {}) {
    if (!this.map || !data || data.length < 2) return;

    //转换为GCJ02坐标
    let path = data.map((item: any) => {
      let [lng, lat] = this.wgs84ToGcj02(item.longitude, item.latitude);
      return [lng, lat];
    })

    //创建起点marker
    let startMarker = new this.AMap.Marker({
      position: path[0],
      icon: options.startIcon || 'https://webapi.amap.com/theme/v1.3/markers/n/mark_b.png',
    })
    this.map.add(startMarker);
    this.overlays.push(startMarker);

    //创建动画marker
    let marker = new this.AMap.Marker({
      position: path[0],
      icon: options.icon || 'https://webapi.amap.com/theme/v1.3/markers/n/mark_r.png',
    })
    this.map.add(marker);
    this.overlays.push(marker);

    //绘制路径
    let polyline = new this.AMap.Polyline({
      path: path,
      strokeColor: options.strokeColor || "#2b8cbe",
      strokeWeight: options.strokeWeight || 4
    })
    this.map.add(polyline);
    this.overlays.push(polyline);

    //执行动画
    marker.moveAlong(path, {
      duration: options.duration || 5000,
      trail: options.trail || false,
    })

    return marker;
  }

  // ==================== 地图控制接口 ====================

  /**
   * 清空所有覆盖物
   */
  clearOverlays() {
    if (!this.map) return;

    this.overlays.forEach(overlay => {
      this.map.remove(overlay);
    });
    this.overlays = [];
  }

  /**
   * 设置地图中心点
   * @param longitude 经度 - WGS84坐标
   * @param latitude 纬度 - WGS84坐标
   * @param zoom 缩放级别（可选）
   */
  setCenter(longitude: number, latitude: number, zoom?: number) {
    if (!this.map) return;

    //WGS84转GCJ02
    let [lng, lat] = this.wgs84ToGcj02(longitude, latitude);
    this.map.setCenter([lng, lat]);
    if (zoom !== undefined) {
      this.map.setZoom(zoom);
    }
  }

  /**
   * 设置缩放级别
   * @param level 缩放级别
   */
  setZoom(level: number) {
    if (!this.map) return;
    this.map.setZoom(level);
  }

  /**
   * 视野自适应
   * @param overlays 指定覆盖物数组（可选）
   */
  fitBounds(overlays?: any[]) {
    if (!this.map) return;

    if (overlays && overlays.length > 0) {
      this.map.setFitView(overlays);
    } else {
      this.map.setFitView();
    }
  }

  /**
   * 获取当前地图中心点
   * @returns 坐标数组 [经度, 纬度] - WGS84坐标
   */
  getCenter(): number[] | null {
    if (!this.map) return null;

    let center = this.map.getCenter();
    //GCJ02转WGS84
    return this.gcj02ToWgs84(center.lng, center.lat);
  }

  /**
   * 获取当前缩放级别
   * @returns 缩放级别
   */
  getZoom(): number | null {
    if (!this.map) return null;
    return this.map.getZoom();
  }

  // ==================== 位置搜索接口 ====================

  /**
   * 地理编码（地址转坐标）
   * @param address 地址字符串
   * @param city 城市（可选）
   * @returns Promise 解析为 {longitude, latitude} - WGS84坐标
   */
  async geocodeAddress(address: string, city?: string): Promise<{ longitude: number, latitude: number } | null> {
    if (!this.geocoder) return null;

    return new Promise((resolve) => {
      this.geocoder.getLocation(address, (status: string, result: any) => {
        if (status === 'complete' && result.geocodes.length > 0) {
          let location = result.geocodes[0].location;
          //GCJ02转WGS84
          let [lng, lat] = this.gcj02ToWgs84(location.lng, location.lat);
          resolve({ longitude: lng, latitude: lat });
        } else {
          console.error('地理编码失败:', status);
          resolve(null);
        }
      });
    });
  }

  /**
   * 逆地理编码（坐标转地址）
   * @param longitude 经度 - WGS84坐标
   * @param latitude 纬度 - WGS84坐标
   * @returns Promise 解析为地址字符串
   */
  async reverseGeocodeLocation(longitude: number, latitude: number): Promise<string | null> {
    if (!this.geocoder) return null;

    //WGS84转GCJ02
    let [lng, lat] = this.wgs84ToGcj02(longitude, latitude);

    return new Promise((resolve) => {
      this.geocoder.getAddress([lng, lat], (status: string, result: any) => {
        if (status === 'complete' && result.regeocode) {
          resolve(result.regeocode.formattedAddress);
        } else {
          console.error('逆地理编码失败:', status);
          resolve(null);
        }
      });
    });
  }

  /**
   * 获取当前设备位置
   * @returns Promise 解析为 {longitude, latitude} - WGS84坐标
   */
  async getCurrentPosition(): Promise<{ longitude: number, latitude: number } | null> {
    if (!this.AMap) return null;

    return new Promise((resolve) => {
      this.AMap.plugin('AMap.Geolocation', () => {
        let geolocation = new this.AMap.Geolocation({
          enableHighAccuracy: true,
          timeout: 10000,
        });

        geolocation.getCurrentPosition((status: string, result: any) => {
          if (status === 'complete') {
            //GCJ02转WGS84
            let [lng, lat] = this.gcj02ToWgs84(result.position.lng, result.position.lat);
            resolve({
              longitude: lng,
              latitude: lat
            });
          } else {
            console.error('获取位置失败:', result.message);
            resolve(null);
          }
        });
      });
    });
  }

  // ==================== 坐标转换接口 ====================

  /**
   * WGS84转GCJ02（GPS坐标转高德/谷歌坐标）
   * @param longitude WGS84经度
   * @param latitude WGS84纬度
   * @returns GCJ02坐标 [经度, 纬度]
   */
  wgs84ToGcj02(longitude: number, latitude: number): number[] {
    return convert.wgs84ToGcj02(longitude, latitude);
  }

  /**
   * GCJ02转WGS84（高德/谷歌坐标转GPS坐标）
   * @param longitude GCJ02经度
   * @param latitude GCJ02纬度
   * @returns WGS84坐标 [经度, 纬度]
   */
  gcj02ToWgs84(longitude: number, latitude: number): number[] {
    return convert.gcj02ToWgs84(longitude, latitude);
  }

  /**
   * WGS84转BD09（GPS坐标转百度坐标）
   * @param longitude WGS84经度
   * @param latitude WGS84纬度
   * @returns BD09坐标 [经度, 纬度]
   */
  wgs84ToBd09(longitude: number, latitude: number): number[] {
    return convert.wgs84ToBd09(longitude, latitude);
  }

  /**
   * BD09转WGS84（百度坐标转GPS坐标）
   * @param longitude BD09经度
   * @param latitude BD09纬度
   * @returns WGS84坐标 [经度, 纬度]
   */
  bd09ToWgs84(longitude: number, latitude: number): number[] {
    return convert.bd09ToWgs84(longitude, latitude);
  }

  /**
   * GCJ02转BD09（高德坐标转百度坐标）
   * @param longitude GCJ02经度
   * @param latitude GCJ02纬度
   * @returns BD09坐标 [经度, 纬度]
   */
  gcj02ToBd09(longitude: number, latitude: number): number[] {
    return convert.gcj02ToBd09(longitude, latitude);
  }

  /**
   * BD09转GCJ02（百度坐标转高德坐标）
   * @param longitude BD09经度
   * @param latitude BD09纬度
   * @returns GCJ02坐标 [经度, 纬度]
   */
  bd09ToGcj02(longitude: number, latitude: number): number[] {
    return convert.bd09ToGcj02(longitude, latitude);
  }

  // ==================== GeoHash接口 ====================

  /**
   * 坐标转GeoHash
   * @param longitude 经度
   * @param latitude 纬度
   * @param precision 精度（可选，默认12）
   * @returns GeoHash字符串
   */
  encodeGeoHash(longitude: number, latitude: number, precision: number = 12): string {
    return encodeBase32(latitude, longitude);
  }

  // ==================== 多边形编辑器接口 ====================

  /**
   * 初始化多边形编辑器
   */
  initPolygonEditor() {
    console.log("init polygon editor")
    this.polygonEditor = new this.AMap.PolygonEditor(this.map);

    this.polygonEditor.on('add', (data: any) => {
      console.log(data);
      var polygon = data.target;
      this.polygonEditor.addAdsorbPolygons(polygon);

      polygon.on('dblclick', () => {
        this.polygonEditor.setTarget(polygon);
        this.polygonEditor.open();
      })
    })
  }

  /**
   * 开始创建/编辑多边形
   */
  startPolygonEdit() {
    if (!this.polygonEditor) {
      this.initPolygonEditor();
    }
    this.polygonEditor.close();
    this.polygonEditor.setTarget();
    this.polygonEditor.open();
  }

  /**
   * 停止多边形编辑
   */
  stopPolygonEdit() {
    if (this.polygonEditor) {
      this.polygonEditor.close();
    }
  }

  /**
   * 获取当前编辑多边形的路径
   * @returns 多边形顶点数组 [ [lng1, lat1], [lng2, lat2], ... ] - WGS84坐标
   */
  getPolygonPath(): number[][] | null {
    if (!this.polygonEditor) return null;

    let polygon = this.polygonEditor.getTarget();
    if (!polygon) return null;

    let path = polygon.getPath();
    if (!path || path.length < 3) return null;

    //GCJ02转WGS84
    return path.map((point: any) => {
      const lng = typeof point === 'object' && 'lng' in point ? point.lng : point[0];
      const lat = typeof point === 'object' && 'lat' in point ? point.lat : point[1];
      return this.gcj02ToWgs84(lng, lat);
    });
  }

  /**
   * 获取当前编辑的多边形数据
   * @returns {points: number[][], id?: number} | null - WGS84坐标
   */
  getCurrentPolygon(): { points: number[][], id?: number } | null {
    if (!this.polygonEditor) return null;

    let polygon = this.polygonEditor.getTarget();
    if (!polygon) return null;

    let path = polygon.getPath();
    if (!path || path.length < 3) return null;

    //GCJ02转WGS84
    let points = path.map((point: any) => {
      const lng = typeof point === 'object' && 'lng' in point ? point.lng : point[0];
      const lat = typeof point === 'object' && 'lat' in point ? point.lat : point[1];
      return this.gcj02ToWgs84(lng, lat);
    });

    let extData = polygon.getExtData();
    return {
      points: points,
      id: extData?.id
    };
  }

  // ==================== 生命周期 ====================
  override ngOnDestroy() {
    super.ngOnDestroy();
    
    this.stopPolygonEdit();
    this.clearOverlays();
    if (this.map) {
      this.map.destroy();
      this.map = null;
    }
  }
}