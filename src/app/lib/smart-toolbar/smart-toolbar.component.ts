import {Component, EventEmitter, Input, Output} from '@angular/core';
import {FormBuilder, FormGroup, FormsModule, ReactiveFormsModule} from '@angular/forms';
import {CommonModule} from '@angular/common';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {NzIconDirective} from 'ng-zorro-antd/icon';
import {NzInputDirective, NzInputGroupComponent} from 'ng-zorro-antd/input';
import {NzSpaceModule} from 'ng-zorro-antd/space';
import {NzAutocompleteComponent, NzAutocompleteTriggerDirective} from 'ng-zorro-antd/auto-complete';
import {NzCheckboxComponent} from 'ng-zorro-antd/checkbox';
import {NzDatePickerComponent} from 'ng-zorro-antd/date-picker';
import {NzInputNumberComponent} from 'ng-zorro-antd/input-number';
import {NzRadioComponent, NzRadioGroupComponent} from 'ng-zorro-antd/radio';
import {NzSelectComponent} from 'ng-zorro-antd/select';
import {NzSliderComponent} from 'ng-zorro-antd/slider';
import {NzSwitchComponent} from 'ng-zorro-antd/switch';
import {NzTimePickerComponent} from 'ng-zorro-antd/time-picker';
import {NzTreeSelectComponent} from 'ng-zorro-antd/tree-select';
import {NzUploadChangeParam, NzUploadComponent} from 'ng-zorro-antd/upload';
import {createControl, SmartField} from '../smart-editor/smart-editor.component';
import {NzDropDownDirective, NzDropdownMenuComponent} from 'ng-zorro-antd/dropdown';
import {NzMenuDirective, NzMenuItemComponent} from 'ng-zorro-antd/menu';
import {SmartAction} from '../smart-table/smart-table.component';
import {Router} from '@angular/router';
import {NzModalService} from 'ng-zorro-antd/modal';
import {SmartRequestService} from '../smart-request.service';
import {NzFlexDirective} from 'ng-zorro-antd/flex';
import {NzPopconfirmDirective} from 'ng-zorro-antd/popconfirm';

@Component({
  selector: 'smart-toolbar',
  standalone: true,
  imports: [
    FormsModule,
    CommonModule,
    NzSpaceModule,
    NzButtonComponent,
    NzIconDirective,
    NzInputDirective,
    NzAutocompleteComponent,
    NzSelectComponent,
    NzTreeSelectComponent,
    NzSwitchComponent,
    NzCheckboxComponent,
    NzSliderComponent,
    NzDatePickerComponent,
    NzTimePickerComponent,
    NzInputGroupComponent,
    NzUploadComponent,
    NzAutocompleteTriggerDirective,
    ReactiveFormsModule,
    NzInputNumberComponent,
    NzRadioComponent,
    NzRadioGroupComponent,
    NzDropDownDirective,
    NzDropdownMenuComponent,
    NzMenuDirective,
    NzMenuItemComponent,
    NzFlexDirective,
    NzPopconfirmDirective,
  ],
  templateUrl: './smart-toolbar.component.html',
  styleUrl: './smart-toolbar.component.scss'
})
export class SmartToolbarComponent {
  @Output() change = new EventEmitter<any>();
  @Output() action = new EventEmitter<SmartAction>();

  group: FormGroup = new FormGroup({})
  _fields: SmartField[] = []
  _values: any = {}
  empty: any = []

  @Input() set fields(fs: SmartField[]) {
    //console.log("[SmartToolbar] set fields", fs)
    if (fs && fs.length) {
      setTimeout(() => {
        this._fields = fs
        //this.group = this.build()
        //this.group.valueChanges.subscribe(res => this.change.emit(res))
        this.build()
      }, 50)
    }
  }

  get fields() {
    return this._fields
  }

  get values() {
    return this._values
  }

  @Input() set values(values: any) {
    //console.log("[SmartToolbar] set values", values)
    this._values = values
    if (this._fields && this._fields.length) {
      setTimeout(() => {
        this.build()
      }, 50)
    }
  }

  build() {
    //console.log("[SmartToolbar] build", fields, values)
    const values = this.values || {}

    let fs: any = {}
    this.fields?.forEach(f => {
      if (f.type == "button" || f.type == "dropdown")
        return

      //创建组件
      fs[f.key] = createControl(f, values[f.key])

      //订阅变化
      if (f.change)
        fs[f.key].valueChanges.subscribe((res: any) => f.change?.(res))
    })
    this.group = this.fb.group(fs)
    this.group.valueChanges.subscribe((res: any) => this.change.emit(res))
  }

  get valid() {
    return this.group.valid
  }

  get value() {
    return this.group.value
  }

  constructor(private fb: FormBuilder, private router: Router, private ms: NzModalService, private request: SmartRequestService) {
  }

  ngOnInit(): void {
    if (this._fields && this._fields.length) {
      this.build()
    }
  }

  handleUpload(key: string, $event: NzUploadChangeParam) {
    if ($event.type == 'success') {
      this.group.patchValue({[key]: $event.file.response.data[0]})
      //control.setValue($event.file.response.data[0])
    }
  }

  handleUploadImages(key: string, $event: NzUploadChangeParam) {
    let paths: any = []
    $event.fileList.forEach(file => {
      if (file.response?.data?.[0])
        paths.push(file.response.data[0])
    })
    this.group.patchValue({[key]: paths})
    //control.setValue(paths)
  }


  protected readonly Infinity = Infinity;

}
