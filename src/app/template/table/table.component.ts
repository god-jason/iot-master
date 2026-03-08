import { Component, ViewChild } from '@angular/core';
import { ParamSearch, SmartTableComponent } from '../../lib/smart-table/smart-table.component';
import { isFunction } from 'rxjs/internal/util/isFunction';
import { NzCardComponent } from 'ng-zorro-antd/card';
import { CommonModule } from '@angular/common';
import { NzSpinComponent } from 'ng-zorro-antd/spin';
import { NzButtonComponent } from 'ng-zorro-antd/button';
import { NzIconDirective } from 'ng-zorro-antd/icon';
import { SmartToolbarComponent } from '../../lib/smart-toolbar/smart-toolbar.component';
import { TemplateBase } from '../template-base.component';
import { TableContent } from '../template';
import { LinkReplaceParams } from '../../lib/utils';


@Component({
  selector: 'app-table',
  imports: [
    SmartTableComponent,
    NzCardComponent,
    CommonModule,
    NzSpinComponent,
    NzButtonComponent,
    NzIconDirective,
    SmartToolbarComponent,
  ],
  templateUrl: './table.component.html',
  standalone: true,
  styleUrl: './table.component.scss',
  inputs: ['page', 'content', 'params', 'data', 'isChild']
})
export class TableComponent extends TemplateBase {
  @ViewChild("toolbar", { static: false }) toolbar!: SmartToolbarComponent;
  @ViewChild("table", { static: false }) table!: SmartTableComponent;

  toolbarValue = {}

  total = 0

  $event!: ParamSearch // = {filter: {}}

  filter = {}
  keyword = ""

  override load($event?: ParamSearch) {
    if (!$event && !this.$event) return //避免重复调用

    console.log("[table] search", this.page)
    const content = this.content as TableContent
    if (!content) return

    //默认用上次搜索
    if (!$event) $event = this.$event
    else this.$event = $event

    //继承条件
    Object.assign($event.filter, this.filter)

    //关键字
    $event.filter["$or"] = {}
    content.keywords?.forEach((key) => {
      if (this.keyword)
        $event.filter["$or"][key] = "%"+this.keyword+"%"
    })

    //关联查询
    $event.joins = content.joins || []

    //搜索
    if (typeof content.search == "string" && content.search.length > 0) {
      try {
        content.search = new Function("param", "request", content.search)
      } catch (e) {
        console.error(e)
      }
    }
    if (isFunction(content.search)) {
      this.loading = true
      content.search($event, this.request).then((res: any) => {
        this.data = res.data || []
        this.total = res.total || res.data?.length || 0
      }).finally(() => {
        this.loading = false
      })
    } else if (content.search_api) {
      this.loading = true
      let url = LinkReplaceParams(content.search_api, this.params);
      this.request.post(url, $event).subscribe(res => {
        if (res.error) return
        this.data = res.data || []
        this.total = res.total || res.data?.length || 0
      }).add(() => {
        this.loading = false
      })
    } else {
      //调用 load_api
      super.load()
    }
  }

}
