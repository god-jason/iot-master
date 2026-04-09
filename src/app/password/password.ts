import {Component, ViewChild} from '@angular/core';
import {SmartEditorComponent, SmartField} from '../lib/smart-editor/smart-editor.component';
import {Router} from '@angular/router';
import {NzNotificationService} from 'ng-zorro-antd/notification';
import {SmartRequestService} from '../lib/smart-request.service';
import {UserService} from '../user.service';
import {Md5} from 'ts-md5';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzButtonComponent} from 'ng-zorro-antd/button';

@Component({
  selector: 'app-password',
  standalone: true,
  imports: [
    SmartEditorComponent,
    NzCardComponent,
    NzButtonComponent
  ],
  templateUrl: './password.html',
  styleUrl: './password.scss',
})
export class PasswordComponent {

  fields: SmartField[] = [
    {key: 'old', type: 'password', label: '旧密码', required: true},
    {key: 'new', type: 'password', label: '新密码', required: true},
  ]

  @ViewChild("editor", {static: true}) editor!: SmartEditorComponent;

  constructor(private router: Router,
              private ns: NzNotificationService,
              private request: SmartRequestService,
              private us: UserService,
  ) {
  }

  submit() {

    if (!this.editor.valid) {
      this.ns.error("错误", "无效密码")
      return
    }

    let obj = this.editor.value
    this.request.post("password", {
      old: Md5.hashStr(obj.old),
      new: Md5.hashStr(obj.new),
    }).subscribe(res => {
      console.log("password", res)
      if (res.error) {
        return
      }
      this.ns.success("提示", "修改成功，请重新登录")

      localStorage.removeItem("token")
      this.router.navigateByUrl('/login')
    })
  }
}
