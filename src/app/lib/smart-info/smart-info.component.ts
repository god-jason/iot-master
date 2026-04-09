import {Component, EventEmitter, Input, Output} from '@angular/core';
import {CommonModule} from "@angular/common";
import {NzDescriptionsModule} from "ng-zorro-antd/descriptions";
import {NzProgressComponent} from "ng-zorro-antd/progress";
import {NzTagComponent} from "ng-zorro-antd/tag";
import {SmartAction} from '../smart-table/smart-table.component';
import {NzModalModule} from 'ng-zorro-antd/modal';
import {NzBytesPipe} from 'ng-zorro-antd/pipes';
import {NzAvatarComponent} from 'ng-zorro-antd/avatar';


export interface SmartInfoItem {
  key: string
  label: string
  type?: string
  format?: string
  span?: number
  action?: SmartAction
  options?: { [p: string | number]: any }

  //仅限管理员
  admin?: boolean
  //仅限非管理员
  not_admin?: boolean
}

@Component({
  selector: 'smart-info',
  standalone: true,
  imports: [
    CommonModule,
    NzDescriptionsModule,
    NzProgressComponent,
    NzTagComponent,
    NzModalModule,
    NzBytesPipe,
    NzAvatarComponent,
  ],
  templateUrl: './smart-info.component.html',
  styleUrl: './smart-info.component.scss'
})
export class SmartInfoComponent {
  @Input() title: string = '';
  _fields: SmartInfoItem[] = []
  @Input() value: any = {}
  @Output() action = new EventEmitter<SmartAction>();
  @Input() user:any = {}

  @Input() set fields(fs: SmartInfoItem[]) {
    this._fields = fs.filter(f=>{
      //管理员
      if (f.admin)
        return this.user?.admin
      //非管理员
      if (f.not_admin)
        return !(this.user?.admin)
      return true
    })
  }
  get fields() {
    return this._fields
  }

  constructor() {
  }

  protected readonly parseInt = parseInt;
}
