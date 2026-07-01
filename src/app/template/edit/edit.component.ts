import {Component, inject, ViewChild} from '@angular/core';
import {isFunction} from 'rxjs/internal/util/isFunction';
import {SmartEditorComponent} from '../../lib/smart-editor/smart-editor.component';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {NzSkeletonComponent} from 'ng-zorro-antd/skeleton';
import {NzIconDirective} from 'ng-zorro-antd/icon';
import {SmartToolbarComponent} from '../../lib/smart-toolbar/smart-toolbar.component';

import {NzModalRef} from 'ng-zorro-antd/modal';
import {TemplateBase} from '../template-base.component';
import {EditContent} from '../template';
import {LinkReplaceParams} from '../../lib/utils';
import {CommonModule} from '@angular/common';


@Component({
  selector: 'app-edit',
  imports: [
    CommonModule,
    SmartEditorComponent,
    NzButtonComponent,
    NzSkeletonComponent,
    NzIconDirective,
    SmartToolbarComponent
],
  templateUrl: './edit.component.html',
  standalone: true,
  styleUrl: './edit.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class EditComponent extends TemplateBase {
  modalRef = inject(NzModalRef, {optional: true})

  @ViewChild("toolbar", {static: false}) toolbar!: SmartToolbarComponent;
  @ViewChild("editor", {static: false}) editor!: SmartEditorComponent;
  toolbarValue = {}

  submitting = false;

  submit() {
    if (this.submitting) return
    console.log("[edit] submit", this.page)
    const content = this.content as EditContent
    if (!content) return

    if (typeof content.submit == "string" && content.submit.length > 0) {
      try {
        content.submit = new Function("data", content.submit)
      } catch (e) {
        console.error(e)
      }
    }
    if (isFunction(content.submit)) {
      //this.submitting = true
      content.submit.call(this, this.editor.value)
      //   .then((res: any) => {
      //   //this.data = res;
      //   //this.ns.success("提示", "提交成功")
      // }).finally(() => {
      //   this.submitting = false
      // })
    } else if (content.submit_api) {
      this.submitting = true
      let url = LinkReplaceParams(content.submit_api, this.params);
      this.request.post(url, this.editor.value).subscribe(res => {
        if (res.error) return
        //this.data = res.data
        //this.ns.success("提示", "提交成功")
        //Object.assign(this.data, res.data)

        //处理提交成功
        if (typeof content.submit_success == "string" && content.submit_success.length > 0) {
          try {
            content.submit_success = new Function("data", content.submit_success)
          } catch (e) {
            console.error(e)
          }
        }
        if (isFunction(content.submit_success)) {
          content.submit_success.call(this, res.data)
        }
        
        // 提交成功后，显示成功提示
        this.notification.success("提示", "提交成功")

        //关闭弹窗
        if (this.modalRef && !this.isChild)
          this.modalRef.close();
      }).add(() => {
        this.submitting = false
      })
    }
  }


}