import {Component, ViewChild} from '@angular/core';
import {SmartRequestService} from '../lib/smart-request.service';
import {ActivatedRoute} from '@angular/router';
import {SmartEditorComponent} from '../lib/smart-editor/smart-editor.component';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {NzNotificationService} from 'ng-zorro-antd/notification';
import {Title} from '@angular/platform-browser';

@Component({
  selector: 'app-setting',
  standalone: true,
  imports: [
    SmartEditorComponent,
    NzCardComponent,
    NzButtonComponent
  ],
  templateUrl: './setting.component.html',
  styleUrl: './setting.component.scss'
})
export class SettingComponent {
  module = ""
  form: any = {}
  data: any = {}

  @ViewChild("editor", {static: true}) editor!: SmartEditorComponent;


  constructor(protected request: SmartRequestService,
              protected route: ActivatedRoute,
              protected ns: NzNotificationService,
              protected title: Title) {

    route.params.subscribe(params => {
      this.module = params['module']
      this.load()
    })
    //this.load()
  }

  load() {
    this.request.get("setting/" + this.module).subscribe((res) => {
      if (res.error) return
      this.data = res.data
    })
    this.request.get("setting/" + this.module + "/form").subscribe((res) => {
      if (res.error) return
      this.form = res.data
      this.title.setTitle("设置 " + this.form.title)
    })
  }

  submit() {
    this.request.post("setting/" + this.module, this.editor.value).subscribe((res) => {
      this.ns.success("提示", "保存成功")
    })
  }


}
