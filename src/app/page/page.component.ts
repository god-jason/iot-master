import {
  AfterViewInit,
  Component,
  ComponentRef,
  inject,
  Input, OnDestroy,
  ViewChild,
  ViewChildren,
  ViewContainerRef
} from '@angular/core';
import {SmartRequestService} from '../lib/smart-request.service';
import {ActivatedRoute, NavigationEnd, Params, Router, RouterLink} from '@angular/router';
import {NzSpinComponent} from 'ng-zorro-antd/spin';
import {Title} from '@angular/platform-browser';
import {isFunction} from 'rxjs/internal/util/isFunction';
import {NzGridModule} from 'ng-zorro-antd/grid';
import {NZ_MODAL_DATA, NzModalModule} from 'ng-zorro-antd/modal';
import {PageContent} from '../template/template';
import {ObjectDeepCompare} from '../lib/utils';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {NzResultComponent} from 'ng-zorro-antd/result';
import {NzTabsModule} from 'ng-zorro-antd/tabs';
import {Subscription} from "rxjs";
import {CommonModule} from '@angular/common';
import {SmartCardComponent} from '../lib/smart-card/smart-card.component';
import {NzIconModule} from 'ng-zorro-antd/icon';

@Component({
  selector: 'app-page',
  imports: [
    CommonModule,
    NzIconModule,
    SmartCardComponent,
    NzSpinComponent,
    NzButtonComponent,
    NzResultComponent,
    NzGridModule,
    RouterLink,
    NzTabsModule,
    NzModalModule,
  ],
  templateUrl: './page.component.html',
  standalone: true,
  styleUrl: './page.component.scss',
})
export class PageComponent implements AfterViewInit, OnDestroy{
  @Input() page?: string
  @Input() content?: PageContent
  @Input() params?: Params

  @Input() isChild = false

  error = ''

  nzModalData: any = inject(NZ_MODAL_DATA, {optional: true});

  componentRef!: ComponentRef<any>

  @ViewChildren(PageComponent) children!: PageComponent[]

  private routerSub?: Subscription;

  constructor(protected request: SmartRequestService,
              protected route: ActivatedRoute,
              protected router: Router,
              protected title: Title
  ) {
    //优先使用弹窗参数
    if (this.nzModalData) {
      this.page = this.nzModalData.page;
      this.params = this.nzModalData.params;
    } else {
      //this.page = route.snapshot.params['page'];
      this.page = location.pathname.substring(6)
      this.params = route.snapshot.queryParams;
    }
    //console.log("page constructor", this.page, this.params)
  }

  ngOnDestroy() {
    //取消路由事件订阅
    this.routerSub?.unsubscribe()
  }

  ngAfterViewInit(): void {
    if (this.content) {
      //this.ts.setTitle(this.content.title);
      this.build()
    } else {
      if (this.page) this.load_page()

      //弹窗之外，需要监听路由参数
      if (!this.nzModalData && !this.isChild) {
        this.routerSub = this.router.events.subscribe(event=>{
          if (event instanceof NavigationEnd) {
            const page = location.pathname.substring(6)
            //console.log("[page] NavigationEnd:", page);
            if (this.page == page) return
            //console.log("[page] page change:", page);
            this.page = page
            this.load_page()
          }
        })

        // this.route.params.subscribe(params => {
        //   const page = location.pathname.substring(6)
        //   if (this.page == page) return
        //
        //   console.log("[page] page change:", page);
        //
        //   this.page = page
        //   this.load_page()
        // })
        // this.route.params.subscribe(params => {
        //   if (this.page == params['page']) return
        //
        //   console.log("[page] page change")
        //
        //   this.page = params['page'];
        //   this.load_page()
        //
        //   this.content = undefined
        // })
        this.route.queryParams.subscribe(params => {
          if (ObjectDeepCompare(params, this.params)) return
          //console.log("[page] query change")
          this.params = params;
          //this.load()
          //this.load_page() //重新加载
        })
      }
    }
  }

  load_page() {
    //console.log("[page] loadPage", this.page)
    this.error = ''

    this.content = undefined //清空页面

    let url = "page/" + this.page
    this.request.get(url, undefined, {observe: 'response', responseType: "text"}).subscribe({
      next: (res) => {
        try {
          // 解析 contenttype 类型，json 直接解决，js 执行 newFunction，并调用
          const contentType = res.headers.get('Content-Type');
          if (contentType?.includes('application/javascript')) {
            try {
              const jsCode = res.body;
              const fn = new Function(jsCode as string);
              this.content = fn();
            } catch (e) {
              console.error('JS 解析错误:', e);
              this.error = '页面解析失败：' + e;
            }
          } else {
            // JSON 格式直接赋值
            try {
              this.content = JSON.parse(res.body as string);
            } catch (e) {
              console.error('JSON 解析错误:', e);
              this.error = 'JSON 解析失败：' + e;
            }
          }

          if (this.content?.title && !this.isChild)
            this.title.setTitle(this.content.title);
          this.build()
        } catch (e) {
          console.error('页面处理错误:', e);
          this.error = '页面处理失败：' + e;
        }
      },
      error: (err) => {
        console.error('页面加载错误:', err);
        this.error = '页面加载失败：' + (err.message || err.statusText || err);
      }
    });
  }

  // calc_params(c: ChildPage | TabPage): any {
  //   //这里会反复地调用， 所以缓存一下
  //   if (c.params) return c.params;
  //
  //   console.log("calc_params", c, this)
  //
  //   if (isFunction(c.params_func)) {
  //     try {
  //       //@ts-ignore
  //       c.params = c.params_func.call(this, this.params)
  //     } catch (e) {
  //       console.error(e)
  //       c.params = {}
  //     }
  //   }
  //
  //   if (!c.params) {
  //     c.params = this.params
  //   }
  //
  //   return c.params
  // }

  build() {
    //console.log("[page] build", this.page)

    this.load_component(this.content?.template)

    this.content?.children?.forEach(c => {
      if (isFunction(c.params)) {
        c.params = c.params.call(this, this.params)
      }
    })

    this.content?.tabs?.forEach(c => {
      if (isFunction(c.params)) {
        c.params = c.params.call(this, this.params)
      }
    })
  }

  @ViewChild('container', { read: ViewContainerRef }) container!: ViewContainerRef;
  render_component(cmp: any): void {
    this.componentRef = this.container.createComponent(cmp)
    this.componentRef.setInput("page", this.page)
    this.componentRef.setInput("content", this.content)
    this.componentRef.setInput("params", this.params)
    this.componentRef.setInput("isChild", this.isChild)
    this.componentRef.setInput("pageComponent", this)

    if (this.isChild) {

    }
  }

  load_component(tpl?: string) {
    switch (tpl) {
      case undefined:
      case "":
      case "blank":
        import("../template/blank/blank.component").then(m => this.render_component(m.BlankComponent))
        break;
      case "list":
        import("../template/list/list.component").then(m => this.render_component(m.ListComponent))
        break
      case "statistic":
        import("../template/statistic/statistic.component").then(m => this.render_component(m.StatisticComponent))
        break
      case "chart":
        import("../template/chart/chart.component").then(m => this.render_component(m.ChartComponent))
        break
      case "amap":
        import("../template/amap/amap.component").then(m => this.render_component(m.AmapComponent))
        break
      case "markdown":
        import("../template/markdown/markdown.component").then(m => this.render_component(m.MarkdownComponent))
        break
      case "import":
        import("../template/import/import.component").then(m => this.render_component(m.ImportComponent))
        break
      case "export":
        import("../template/export/export.component").then(m => this.render_component(m.ExportComponent))
        break
      case "log":
        import("../template/log/log.component").then(m => this.render_component(m.LogComponent))
        break
      case "text":
        import("../template/text/text.component").then(m => this.render_component(m.TextComponent))
        break
      case "edit":
        import("../template/edit/edit.component").then(m => this.render_component(m.EditComponent))
        break
      case "detail":
        import("../template/detail/detail.component").then(m => this.render_component(m.DetailComponent))
        break
      case "value":
        import("../template/value/value.component").then(m => this.render_component(m.ValueComponent))
        break
      default:
        break
    }
  }

}
