import {Component, inject} from '@angular/core';
import {Router} from '@angular/router';
import {SmartRequestService} from '../lib/smart-request.service';
import {NzModalModule, NzModalRef, NzModalService} from 'ng-zorro-antd/modal';
import {Title} from '@angular/platform-browser';
import {SmartAction} from '../lib/smart-table/smart-table.component';
import {isFunction} from 'rxjs/internal/util/isFunction';
import {PageContent} from './template';
import {PageComponent} from '../page/page.component';
import {LinkReplaceParams} from '../lib/utils';

import dayjs from 'dayjs'

@Component({
  selector: 'app-template',
  imports: [
    NzModalModule,
  ],
  template: '',
  standalone: true,
  inputs: ['page', 'content', 'params', 'data', 'isChild', 'pageComponent']
})
export class TemplateBase {
  dayjs: any = dayjs //引入dayjs

  request = inject(SmartRequestService)
  modal = inject(NzModalService)
  router = inject(Router)
  title = inject(Title)
  modelRef = inject(NzModalRef, {optional: true})

  pageComponent!: PageComponent //页面容器

  page?: string

  _params?: any
  set params(p: any) {
    this._params = p;
    this.load()
  }

  get params(): any {
    return this._params;
  }

  content?: PageContent
  isChild = false

  data: any = []

  loading = false

  auto_refresh_interval = 0

  constructor() {
    //super();
    //console.log("base constructor", this.dayjs())
  }


  ngAfterViewInit() {
    this.init()
  }

  ngOnInit(): void {
    //this.mount()

  }

  ngOnDestroy(): void {
    this.unmount()
  }

  mount() {
    //@ts-ignore
    this.onMount?.call(this)

    //自动刷新
    if (typeof this.content?.auto_refresh === "number" && this.content.auto_refresh > 0) {
      this.auto_refresh_interval = setInterval(() => this.load(), this.content.auto_refresh * 1000)
    }

    if (typeof this.content?.mount == "string" && this.content.mount.length > 0) {
      try {
        this.content.mount = new Function(this.content.mount)
      } catch (e) {
        console.error(e)
      }
    }
    if (isFunction(this.content?.mount)) {
      this.content?.mount.call(this)
    }
  }

  unmount() {
    //@ts-ignore
    this.onUnmount?.call(this)

    //自动刷新
    if (this.auto_refresh_interval)
      clearInterval(this.auto_refresh_interval)

    if (typeof this.content?.unmount == "string" && this.content.unmount.length > 0) {
      try {
        this.content.unmount = new Function(this.content.unmount)
      } catch (e) {
        console.error(e)
      }
    }
    if (isFunction(this.content?.unmount)) {
      this.content?.unmount.call(this)
    }
  }

  init() {
    //由于外层是page，每次都会把content先加载完成，所以这里content不可能为空

    //如果是input传入，则是作为组件使用
    if (this.content) {
      this.build()
      this.mount()
      this.load()
    } else {
      if (this.page)
        this.load_page()
    }
  }

  load_page() {
    console.log("[base] load page", this.page)
    let url = "page/" + this.page
    this.request.get(url).subscribe((res) => {
      if (res.error) return
      this.content = res
      if (this.content?.title && !this.isChild && !this.modelRef)
        this.title.setTitle(this.content.title);
      this.build()
      this.mount()
      this.load()
    })
  }

  //abstract build(): void
  build() {
    console.log("[base] build")

    //编译成员
    if (this.content?.methods != undefined) {
      Object.keys(this.content.methods).forEach(method => {
        let func = this.content?.methods?.[method]
        if (typeof func == "string") {
          try {
            let fn = new Function(func)
            fn.bind(this)
            //@ts-ignore
            this[method] = fn
          } catch (e) {
            console.error(e)
          }
        } else if (Array.isArray(func)) {
          try {
            let fn = new Function(...func)
            fn.bind(this)
            //@ts-ignore
            this[method] = fn
          } catch (e) {
            console.error(e)
          }
        } else {
          //@ts-ignore
          this[method] = func
        }
      })
    }
  }

  render(data: any) {
    setTimeout(() => {
      this.data = data
    })
  }

  load() {
    console.log("[base] load data", this.page)

    if (!this.content) return

    //初始化数据
    if (this.content.data) {
      //this.data = this.content.data
      this.render(this.content.data)
    }

    //通过api加载数据
    if (this.content.load_api) {
      this.loading = true
      let url = LinkReplaceParams(this.content.load_api, this.params);
      this.request.get(url).subscribe(res => {
        if (res.error) return
        //this.data = res.data
        this.render(res.data)

        //处理提交成功
        if (typeof this.content?.load_success == "string" && this.content.load_success.length > 0) {
          try {
            this.content.load_success = new Function("data", this.content.load_success)
          } catch (e) {
            console.error(e)
          }
        }
        if (isFunction(this.content?.load_success)) {
          this.content?.load_success.call(this, res.data)
        }
      }).add(() => {
        this.loading = false
      })
    }
  }

  navigate(uri: string) {
    this.router.navigateByUrl(uri).then()
  }


  execute(action: SmartAction, data?: any, index?: number) {
    if (!action) return

    let params = this.get_action_params(action, data || this.data, index || 0)

    switch (action.type) {
      case 'link':

        let uri = this.get_action_link(action, data || this.data, index || 0)
        let query = new URLSearchParams(params).toString()
        let url = uri + '?' + query

        if (action.external)
          window.open(url)
        else
          this.router.navigateByUrl(url)
        //this.router.navigate([uri], {queryParams: params})

        break

      case 'script':
        if (typeof action.script == "string" && action.script.length > 0) {
          try {
            action.script = new Function("data", "index", action.script)
          } catch (e) {
            console.error(e)
          }
        }
        if (isFunction(action.script)) {
          action.script.call(this, data || this.data, index || 0)
        }
        break

      case 'page':
        this.router.navigate(["/page/" + action.page], {queryParams: params})
        break

      case 'dialog':
        this.modal.create({
          nzContent: PageComponent,
          nzWidth: "80%",
          nzData: {
            page: action.page,
            params: params
          },
          nzFooter: null,
          //nzCloseIcon: 'close-circle',
          //nzMaskClosable: false
        }).afterClose.subscribe(res => {
          //关闭的回调
          if (typeof action.after_close == "string" && action.after_close.length > 0) {
            try {
              action.after_close = new Function("result", "data", "index", action.after_close)
            } catch (e) {
              console.error(e)
            }
          }
          if (isFunction(action.after_close)) {
            action.after_close.call(this, res, data || this.data, index || 0)
          }
        })
        break

    }
  }


  get_action_link(action: SmartAction, data: any, index: number) {
    if (!action.link) return ""

    // 先进行正则替换
    let link = LinkReplaceParams(action.link, data)

    // 计算函数
    if (typeof action.link_func == "string" && action.link_func.length > 0) {
      try {
        action.link_func = new Function(action.link_func)
      } catch (e) {
        console.error(e)
      }
    }
    if (isFunction(action.link_func)) {
      link = action.link_func.call(this, data, index)
    }
    return link
  }

  get_action_params(action: SmartAction, data: any, index: number): any {
    let params = action.params
    // 计算函数
    if (typeof action.params_func == "string" && action.params_func.length > 0) {
      try {
        action.params_func = new Function('data', action.params_func)
      } catch (e) {
        console.error(e)
      }
    }
    if (isFunction(action.params_func)) {
      params = action.params_func.call(this, data, index)
    }
    return params
  }


  export_json(data: any, filename: string) {
    const jsonData = JSON.stringify(data, null, "\t");
    const blob = new Blob([jsonData], {type: 'application/json'});
    const url = URL.createObjectURL(blob);

    const a = document.createElement('a');
    a.href = url;
    a.download = filename + dayjs().format("-YYYYMMDDHHmmss") + ".json";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  import_json(): Promise<any> {
    return new Promise((resolve, reject) => {
      let input = document.createElement('input');//js生成接收文件的DOM
      input.type = "file";
      input.click()
      input.onchange = (e) => {
        //@ts-ignore
        let file = e.target?.files?.[0];
        if (file) {
          const reader = new FileReader();
          reader.onload = function (e) {
            let data = e.target?.result || ''
            //console.log("import_json", data)
            try {
              let obj = JSON.parse(data.toString());
              resolve(obj);
            } catch (e) {
              reject(e)
            }
          };
          reader.readAsText(file); // 读取文件内容为文本
        }
      }
    })
  }

}
