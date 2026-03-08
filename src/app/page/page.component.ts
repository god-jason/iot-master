import {
  Component, ComponentRef,
  inject,
  Input, Optional,
  ViewChild, ViewChildren,
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

@Component({
  selector: 'app-page',
  imports: [
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
export class PageComponent {
  @Input() page?: string
  @Input() content?: PageContent
  @Input() params?: Params

  @Input() isChild = false

  error = ''

  nzModalData: any = inject(NZ_MODAL_DATA, {optional: true});

  componentRef!: ComponentRef<any>

  @ViewChildren(PageComponent) children!: PageComponent[]

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
    console.log("page constructor", this.page, this.params)
  }

  ngAfterViewInit(): void {
    if (this.content) {
      //this.ts.setTitle(this.content.title);
      this.build()
    } else {
      if (this.page) this.load_page()

      //弹窗之外，需要监听路由参数
      if (!this.nzModalData && !this.isChild) {
        this.router.events.subscribe(event=>{
          if (event instanceof NavigationEnd) {
            const page = location.pathname.substring(6)
            //console.log("[page] NavigationEnd:", page);
            if (this.page == page) return
            console.log("[page] page change:", page);
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
          console.log("[page] query change")
          this.params = params;
          //this.load()
          //this.load_page() //重新加载
        })
      }
    }
  }

  load_page() {
    console.log("[page] loadPage", this.page)
    this.error = ''

    this.content = undefined //清空页面

    let url = "page/" + this.page
    this.request.get(url).subscribe((res) => {
      if (res.error) {
        //console.log("load page error", res.error)
        this.error = res.error
        return
      }
      this.content = res
      if (this.content?.title && !this.isChild)
        this.title.setTitle(this.content.title);
      this.build()
    }, (error) => {
      this.error = error
    })
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
    console.log("[page] build", this.page)

    this.load_component(this.content?.template)

    this.content?.children?.forEach(c => {
      if (typeof c.params_func == "string") {
        try {
          //@ts-ignore
          c.params_func = new Function('params', c.params_func as string)
        } catch (e) {
          console.error(e)
        }
      }

      //是不是算的有点早了。。。
      if (isFunction(c.params_func)) {
        c.params = c.params_func(this.params)
      }
    })

    this.content?.tabs?.forEach(c => {
      if (typeof c.params_func == "string") {
        try {
          //@ts-ignore
          c.params_func = new Function('params', c.params_func as string)
        } catch (e) {
          console.error(e)
        }
      }

      //是不是算的有点早了。。。
      if (isFunction(c.params_func)) {
        c.params = c.params_func(this.params)
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
      case "table":
        import("../template/table/table.component").then(m => this.render_component(m.TableComponent))
        //this.render_component(TableComponent)
        break
      case "info":
        import("../template/info/info.component").then(m => this.render_component(m.InfoComponent))
        //this.render_component(InfoComponent)
        break
      case "form":
        import("../template/form/form.component").then(m => this.render_component(m.FormComponent))
        //this.render_component(FormComponent)
        break
      case "statistic":
        import("../template/statistic/statistic.component").then(m => this.render_component(m.StatisticComponent))
        //this.render_component(StatisticComponent)
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
      default:
        break
    }
  }

}
