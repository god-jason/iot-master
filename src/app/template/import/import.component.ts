import {Component, inject} from '@angular/core';
import {NzUploadComponent, NzUploadXHRArgs} from 'ng-zorro-antd/upload';
import {read, utils, writeFile} from 'xlsx';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {Subscription} from 'rxjs';
import {NgForOf, NgIf} from '@angular/common';
import {
  NzTableCellDirective,
  NzTableComponent,
  NzTbodyComponent,
  NzTheadComponent,
  NzTrDirective
} from 'ng-zorro-antd/table';
import {TemplateBase} from '../template-base.component';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzIconDirective} from 'ng-zorro-antd/icon';
import {NzSpinComponent} from 'ng-zorro-antd/spin';
import {ImportContent} from '../template';
import {isFunction} from 'rxjs/internal/util/isFunction';
import {LinkReplaceParams} from '../../lib/utils';
import {NzSelectComponent} from 'ng-zorro-antd/select';
import {FormBuilder, FormControl, FormGroup, ReactiveFormsModule} from '@angular/forms';
import {NzStepComponent, NzStepsComponent} from 'ng-zorro-antd/steps';
import {NzProgressComponent} from 'ng-zorro-antd/progress';
import {NzNotificationService} from 'ng-zorro-antd/notification';
import {NzResultComponent, NzResultStatusType} from 'ng-zorro-antd/result';
import dayjs from 'dayjs';


@Component({
  selector: 'app-import',
  imports: [
    NzUploadComponent,
    NzButtonComponent,
    NgForOf,
    NzTableCellDirective,
    NzTableComponent,
    NzTbodyComponent,
    NzTheadComponent,
    NzTrDirective,
    NzCardComponent,
    NzIconDirective,
    NzSpinComponent,
    NzSelectComponent,
    ReactiveFormsModule,
    NzStepsComponent,
    NzStepComponent,
    NzProgressComponent,
    NzResultComponent,
    NgIf,
  ],
  templateUrl: './import.component.html',
  standalone: true,
  styleUrl: './import.component.scss'
})
export class ImportComponent extends TemplateBase {
  datum: any = []

  values: any[] = []

  succeed: any[] = []
  failed: any[] = []

  percent = 0

  current: number = 0;

  group: FormGroup = new FormGroup([])
  options: any = [{value: -1, label: '-'}]

  fb = inject(FormBuilder)
  ns = inject(NzNotificationService)
  //headers: any = {}

  // onMount() {
  //   console.log("[import]", "onMount")
  //
  //   //this.buildForm()
  // }

  resultStatus: NzResultStatusType = "success"
  resultTitle = '上传完成'
  resultReport = ""


  onUploadRequest = (args: NzUploadXHRArgs): Subscription => {
    console.log("onUploadRequest", args);
    //this.datum = [1,2,3,4,4,]

    const sub = new Subscription()
    const reader = new FileReader()
    reader.onload = () => {
      //args.onSuccess?.(true, args.file, args)
    }
    reader.onerror = err => {
      args.onError?.(err, args.file)
    }
    reader.onprogress = ev => {
      args.onProgress?.(ev, args.file)
    }
    reader.onloadend = (e: Event) => {
      args.onSuccess?.(true, args.file, args)


        let wb = read(reader.result)
        let sheet = wb.Sheets[wb.SheetNames[0]]
        let aoa = utils.sheet_to_json(sheet, {header: 1})
        this.datum = aoa

        this.buildForm()
        this.findValues()
        this.current = 1

      //this.buildForm()
      //this.findValues()
    }
    //@ts-ignore
    reader.readAsArrayBuffer(args.file)
    return sub
  }

  buildForm() {
    const content = this.content as ImportContent
    if (!content) return

    let group: any = {}
    content.columns.forEach(c => {
      group[c.key] = new FormControl(-1) //[-1];
    })
    this.group = this.fb.group(group)
  }

  findValues() {
    const content = this.content as ImportContent
    if (!content) return

    let row: any = this.datum[0] || []

    console.log("row", row)
    let headers: any = {}

    this.options = [{value: -1, label: '-'}]
    row.forEach((h: string, i: number) => {
      let ss = h.split("/")
      ss.forEach((v: string) => {
        let key = v.trim()
        content.columns.forEach(c => {
          if (c.key == key || c.label == key)
            headers[c.key] = i
        })
      })

      this.options.push({value: i, label: h,})
    })

    //console.log("options", this.options)
    console.log("headers", headers)
    // setTimeout(()=>{
    //   this.group.patchValue(headers)
    // }, 100)

    this.group.patchValue(headers)
  }

  downloadTemplate($event: Event) {
    $event.stopPropagation()

    const content = this.content as ImportContent
    if (!content) return

    const aoa = content.columns.map(c => {
      if (c.label) return c.label + '/' + c.key
      else return c.key
    })


      const sheet = utils.aoa_to_sheet([aoa])
      const wb = utils.book_new();
      utils.book_append_sheet(wb, sheet)

      // if(!wb.Props) wb.Props = {};
      // wb.Props.Title = "Template";
      // wb.Props.Author = "Boat";

      //const filename = (this.app||'') + (this.page||'') + "-template.xlsx"
      const filename = [this.page?.replaceAll(/\//g, "-"), "template.xlsx"].filter(i => i).join("-")
      writeFile(wb, filename, {compression: true});
  }

  back() {
    this.current--
  }

  preview() {
    this.current = 2

    this.values = []

    let fields = this.group.value
    for (let i = 1; i < this.datum.length; i++) {
      let obj: any = {}
      let row: any[] = this.datum[i]
      for (const k in fields) {
        const index = fields[k]
        if (index > -1 && index < row.length) {
          obj[k] = row[index]
          //todo 处理日期格式
        }
      }
      this.values.push(obj)
    }
  }

  removeValue(i: number): void {
    this.values.splice(i, 1)
  }


  uploading = false;

  async upload() {
    this.current = 3

    if (this.uploading) return
    console.log("[import] submit", this.page)

    const content = this.content as ImportContent
    if (!content) return

    //编译函数
    if (typeof content.submit == "string" && content.submit.length > 0) {
      try {
        content.submit = new Function("data", content.submit)
      } catch (e) {
        console.error(e)
      }
    }

    //逐一上传
    for (let i = 0; i < this.values.length; i++) {
      let row = this.values[i]
      try {
        let obj = await this.uploadRow(row)

        this.succeed.push(obj)
      } catch (e: any) {
        row["_ERROR_"] = e
        this.failed.push(row)
        //this.ns.error("错误", e)
      }

      this.percent = Math.floor((this.succeed.length + this.failed.length) * 100 / this.values.length)
    }

    //完成
    this.current = 4

    if (this.failed.length > 0) {
      this.resultStatus = "error"
      this.resultTitle = '上传失败'
    }
    this.resultReport = `总计${this.values.length}条，已经上传${this.succeed.length}条，失败${this.failed.length}条`
  }

  async uploadRow(row: any) {
    const content = this.content as ImportContent
    if (!content) return

    if (isFunction(content.submit)) {
      return content.submit.call(this, this.data)
    }

    return new Promise((resolve, reject) => {
      let url = LinkReplaceParams(content.submit_api || '', this.params);
      //这里开始上传
      this.request.post(url, row).subscribe({
        next: (res) => {
          if (res.error) {
            reject(res.error)
            return
          }
          resolve(res.data)
        },
        error: (err) => {
          reject(err)
        }
      })
    })
  }

  finish() {
    const content = this.content as ImportContent
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

  downloadSucceed() {
      const sheet = utils.json_to_sheet(this.succeed)
      const wb = utils.book_new();
      utils.book_append_sheet(wb, sheet)
      const filename = [this.page?.replaceAll(/\//g, "-"), dayjs().format('YYYYMMDDHHmmss'), "succeed.xlsx"].filter(i => i).join("-")
      writeFile(wb, filename, {compression: true});
  }

  downloadFailed() {
      const sheet = utils.json_to_sheet(this.failed)
      const wb = utils.book_new();
      utils.book_append_sheet(wb, sheet)
      const filename = [this.page?.replaceAll(/\//g, "-"), dayjs().format('YYYYMMDDHHmmss'), "failed.xlsx"].filter(i => i).join("-")
      writeFile(wb, filename, {compression: true});
  }
}
