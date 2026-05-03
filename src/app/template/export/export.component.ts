import {Component, inject} from '@angular/core';
import {TemplateBase} from '../template-base.component';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {NzIconDirective} from 'ng-zorro-antd/icon';
import {NzSpinComponent} from 'ng-zorro-antd/spin';
import {FormBuilder, FormControl, FormGroup, ReactiveFormsModule} from '@angular/forms';
import {NzNotificationService} from 'ng-zorro-antd/notification';
import {NzResultComponent, NzResultStatusType} from 'ng-zorro-antd/result';
import {ExportContent} from '../template';
import {utils, writeFile} from 'xlsx';
import {isFunction} from 'rxjs/internal/util/isFunction';
import {NzUploadComponent} from 'ng-zorro-antd/upload';
import {
  NzTableCellDirective,
  NzTableComponent,
  NzTbodyComponent,
  NzTheadComponent,
  NzTrDirective
} from 'ng-zorro-antd/table';
import {NzSelectComponent} from 'ng-zorro-antd/select';
import {NzStepComponent, NzStepsComponent} from 'ng-zorro-antd/steps';
import {NzProgressComponent} from 'ng-zorro-antd/progress';
import {NzCheckboxComponent} from 'ng-zorro-antd/checkbox';
import {LinkReplaceParams} from '../../lib/utils';

@Component({
  selector: 'app-export',
  imports: [
    NzUploadComponent,
    NzButtonComponent,
    NzTableCellDirective,
    NzTableComponent,
    NzTbodyComponent,
    NzTheadComponent,
    NzTrDirective,
    NzCardComponent,
    NzIconDirective,
    NzSpinComponent,
    NzSelectComponent,
    NzCheckboxComponent,
    ReactiveFormsModule,
    NzStepsComponent,
    NzStepComponent,
    NzProgressComponent,
    NzResultComponent
  ],
  templateUrl: './export.component.html',
  standalone: true,
  styleUrl: './export.component.scss'
})
export class ExportComponent extends TemplateBase {

  exporting = false


  percent = 0

  total = 0
  downloaded = 0

  current: number = 0;

  group: FormGroup = new FormGroup([])
  fb = inject(FormBuilder)
  ns = inject(NzNotificationService)

  resultStatus: NzResultStatusType = "success"
  resultTitle = '下载完成'
  resultReport = ""

  constructor() {
    super();

    //this.buildForm()
  }

  override ngOnInit() {
    super.ngOnInit();
    this.buildForm()
  }

  buildForm() {
    const content = this.content as ExportContent
    if (!content) return

    console.log("buildForm")

    let group: any = {}
    content.columns.forEach(c => {
      group[c.key] = new FormControl(true) //[-1];
    })
    this.group = this.fb.group(group)
  }

  async download() {
    //$event.stopPropagation()


    this.current = 1


    //TODO 开始下载
    let filter = JSON.parse(this.params.filter || "{}")


    const content = this.content as ExportContent
    if (!content) return

    let fields: string[] = []

    const aoa = content.columns.filter(c => {
      return this.group.value[c.key]
    }).map(c => {
      fields.push(c.key)
      if (c.label) return c.label + '/' + c.key
      else return c.key
    })


    const sheet = utils.aoa_to_sheet([aoa])
    const wb = utils.book_new();
    utils.book_append_sheet(wb, sheet)

    // if(!wb.Props) wb.Props = {};
    // wb.Props.Title = "Template";
    // wb.Props.Author = "Boat";

    let skip = 0
    let limit = 50

    let datum: any[] = []

    while (true) {
      let data: any = await this.requestBatch(filter, skip, limit)
      skip += limit

      if (data && data.length) {
        datum = datum.concat(data)
      }

      if (!data || data.length < limit)
        break

      this.downloaded += data.length
      this.percent = Math.floor(this.downloaded / this.total * 100)
    }

    this.percent = 100

    if (datum.length > 0) {
      let rows = datum.map((d: any) => fields.map(f => d[f]))
      utils.sheet_add_aoa(sheet, [aoa].concat(rows))
    }

    //const filename = (this.app||'') + (this.page||'') + "-template.xlsx"
    const filename = [this.page?.replaceAll(/\//g, "-"), "export.xlsx"].filter(i => i).join("-")
    writeFile(wb, filename, {compression: true});
  }


  async requestBatch(filter: any, skip: number, limit: number) {
    const content = this.content as ExportContent
    if (!content) return

    let param: any = {
      filter, skip, limit
    }

    //搜索
    if (typeof content.search == "string" && content.search.length > 0) {
      try {
        content.search = new Function("param", "request", content.search)
      } catch (e) {
        console.error(e)
      }
    }

    return new Promise((resolve, reject) => {
      if (isFunction(content.search)) {
        this.loading = true
        content.search(param, this.request).then((res: any) => {
          if (res.error) {
            reject(res.error)
            return
          }
          resolve(res.data)

          this.total = res.total || res.data?.length || 0
        }).finally(() => {
          this.loading = false
        })
      } else if (content.search_api) {
        this.loading = true
        let url = LinkReplaceParams(content.search_api, this.params);
        this.request.post(url, param).subscribe(res => {
          if (res.error) {
            reject(res.error)
            return
          }
          resolve(res.data)

          this.total = res.total || res.data?.length || 0
        }).add(() => {
          this.loading = false
        })
      } else {
        reject("缺少查询接口")
      }
    })
  }


  finish() {
    const content = this.content as ExportContent
    if (!content) return

    if (typeof content.finish == "string" && content.finish.length > 0) {
      try {
        content.finish = new Function("data", content.finish)
      } catch (e) {
        console.error(e)
      }
    }
    if (isFunction(content.finish)) {
      content.finish.call(this, this.data)
    }
  }

}
