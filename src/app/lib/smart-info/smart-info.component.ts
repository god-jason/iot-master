import {Component, EventEmitter, Input, Output} from '@angular/core';
import {CommonModule} from "@angular/common";
import {NzDescriptionsModule} from "ng-zorro-antd/descriptions";
import {NzProgressComponent} from "ng-zorro-antd/progress";
import {NzTagComponent} from "ng-zorro-antd/tag";
import {SmartAction} from '../smart-table/smart-table.component';
import {NzModalModule} from 'ng-zorro-antd/modal';
import {NzBytesPipe} from 'ng-zorro-antd/pipes';
import {NzSwitchComponent} from 'ng-zorro-antd/switch';
import {NzAvatarComponent} from 'ng-zorro-antd/avatar';


export interface SmartInfoItem {
  key: string
  label: string
  type?: string
  format?: string
  span?: number
  action?: SmartAction
  options?: { [p: string | number]: any }
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
  @Input() fields: SmartInfoItem[] = []
  @Input() value: any = {}
  @Output() action = new EventEmitter<SmartAction>();

  constructor() {
  }

  protected readonly parseInt = parseInt;
}
