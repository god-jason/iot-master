import {Component, EventEmitter, Input, OnInit, Output, TemplateRef} from '@angular/core';
import {CommonModule} from "@angular/common";
import {
  FormArray,
  FormBuilder,
  FormControl,
  FormGroup,
  FormsModule,
  ReactiveFormsModule,
  Validators
} from "@angular/forms";
import {NzFormModule} from "ng-zorro-antd/form";
import {NzUploadChangeParam, NzUploadComponent} from "ng-zorro-antd/upload";
import {NzTableModule} from "ng-zorro-antd/table";
import {CdkDrag, CdkDragDrop, CdkDragHandle, CdkDropList, moveItemInArray} from "@angular/cdk/drag-drop";
import {NzInputDirective, NzInputGroupComponent, NzTextareaCountComponent} from "ng-zorro-antd/input";
import {NzButtonComponent} from "ng-zorro-antd/button";
import {NzColorPickerModule} from "ng-zorro-antd/color-picker";
import {NzDatePickerComponent} from "ng-zorro-antd/date-picker";
import {NzInputNumberComponent} from "ng-zorro-antd/input-number";
import {NzSelectComponent} from "ng-zorro-antd/select";
import {NzSliderComponent} from "ng-zorro-antd/slider";
import {NzSwitchComponent} from "ng-zorro-antd/switch";
import {NzTimePickerComponent} from "ng-zorro-antd/time-picker";
import {NzTreeSelectComponent} from "ng-zorro-antd/tree-select";
import {NzIconDirective} from "ng-zorro-antd/icon";
import {NzSpaceModule} from "ng-zorro-antd/space";
import {NzAutocompleteModule} from "ng-zorro-antd/auto-complete";
import {NzCheckboxComponent} from "ng-zorro-antd/checkbox";
import {NzRadioModule} from "ng-zorro-antd/radio";
import {NzRateComponent} from "ng-zorro-antd/rate";
import {SmartAction} from '../smart-table/smart-table.component';


export interface SmartAutoOption {
  label: string
  value: any
}

export interface SmartSelectOption {
  label: string
  value: any
  title?: string
  disabled?: boolean
  hide?: boolean
  key?: string | number
}

export interface SmartTreeOption {
  title: string;
  key: string;
  isLeaf?: boolean;
  selectable?: boolean;
  disabled?: boolean;
  expanded?: boolean;
  children?: SmartTreeOption[];

  [key: string]: any;
}

export interface SmartField {
  key: string
  type: string
  label: string
  default?: any
  placeholder?: string
  tips?: string

  clear?: boolean
  disabled?: boolean
  hidden?: boolean //隐藏？？？

  array?: boolean //数组
  children?: SmartField[]


  required?: boolean //必须
  max?: number
  min?: number
  step?: number

  multiple?: boolean //多选

  auto?: SmartAutoOption[] //自动完成
  options?: SmartSelectOption[] //select参数
  tree?: SmartTreeOption[] //树形选择

  change?: (value: any) => void //监测变化
  change_action?: SmartAction

  time?: boolean //日期控件 显示时间
  time_format?: string


  upload?: string //文件上传

  pattern?: string | RegExp
  validators?: any[];

  template?: TemplateRef<any>

  //按钮相关
  icon?: string //按钮类型
  text?: string
  action?: SmartAction //响应
  confirm?: string //确认提醒

  //菜单相关
  dropdown?: {
    label: string,
    action: SmartAction,
  }[]

  //显示的条件
  condition?: SmartFieldCondition
}

export interface SmartFieldCondition {
  key: string
  type: '=' | '==' | '>' | '>=' | '<' | '<=' | '!=' | '<>' | '~=' | 'in' | 'out'
  value: any
}

function getDefault(field: SmartField): any {
  switch (field.type) {
    case 'text':
      return ''
    case 'password':
      return ''
    case 'number':
      return 0
    case 'slider':
      return 0
    case 'radio':
    case 'select':
      return field.options?.[0]?.value
    case 'tags':
      return []
    case 'color':
      return ''
    case 'checkbox':
    case 'switch':
      return false
    case 'textarea':
      return ''
    case 'date':
      return new Date()
    case 'time':
      return new Date()
    case 'datetime':
      return new Date()
    case 'file':
      return ''
    case 'image':
      return ''
    case 'images':
      return []
    case 'object':
      return {}
    case 'list':
      return []
    case 'table':
      return []
  }
  return ''
}

export function createControl(f: SmartField, value: any = undefined): FormControl {
  let validators: any = [];

  if (f.required)
    validators.push(Validators.required)

  if (f.min !== undefined) {
    if (f.type === "number")
      validators.push(Validators.min(f.min))
    else if (f.type === "text" || f.type === "password")
      validators.push(Validators.minLength(f.min))
  }

  if (f.max !== undefined) {
    if (f.type === "number")
      validators.push(Validators.max(f.max))
    else if (f.type === "text" || f.type === "password")
      validators.push(Validators.maxLength(f.max))
  }

  if (f.pattern && f.type === "text")
    validators.push(Validators.pattern(f.pattern))

  //拼接默认校验器
  if (f.validators)
    validators = validators.concat(f.validators)

  //默认值
  if (value === undefined) {
    if (f.hasOwnProperty('default'))
      value = f.default
    else
      value = getDefault(f)
  }

  return new FormControl(value, validators)
}

@Component({
  selector: 'smart-editor',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    NzFormModule,
    NzTableModule,
    CdkDrag,
    CdkDropList,
    NzInputDirective,
    NzButtonComponent,
    NzColorPickerModule,
    NzDatePickerComponent,
    NzInputGroupComponent,
    NzInputNumberComponent,
    NzSelectComponent,
    NzSliderComponent,
    NzSwitchComponent,
    NzTextareaCountComponent,
    NzTimePickerComponent,
    NzTreeSelectComponent,
    NzUploadComponent,
    CdkDragHandle,
    NzIconDirective,
    NzSpaceModule,
    NzAutocompleteModule,
    NzCheckboxComponent,
    NzRadioModule,
    NzRateComponent,
  ],
  templateUrl: './smart-editor.component.html',
  styleUrl: './smart-editor.component.scss',
})
export class SmartEditorComponent implements OnInit {
  @Output() change = new EventEmitter<any>();
  @Output() action = new EventEmitter<SmartAction>();

  group: FormGroup = new FormGroup({})

  _fields: SmartField[] = []
  _values: any = {}

  empty: any = []


  @Input() set fields(fs: SmartField[]) {
    //console.log("[SmartEditor] set fields", fs)
    if (fs && fs.length) {
      setTimeout(() => {
        this._fields = fs
        this.group = this.build(this._fields, this._values)
        this.group.valueChanges.subscribe(res => this.change.emit(res))
      }, 50)
    }
  }

  get fields() {
    return this._fields
  }


  @Input() set values(values: any) {
    //console.log("[SmartEditor] set values", values)
    this._values = values
    if (this._fields && this._fields.length) {
      setTimeout(() => {
        this.group = this.build(this._fields, this._values)
        this.group.valueChanges.subscribe(res => this.change.emit(res))
      }, 50)
    }
  }

  get values() {
    return this._values
  }

  rebuild() {
    if (this._fields && this._fields.length) {
      setTimeout(() => {
        //this.group = this.build(this._fields, this.group.value)
        let value = Object.assign({}, this._values, this.group.value)
        this.group = this.build(this._fields, value)
        this.group.valueChanges.subscribe(res => this.change.emit(res))
      }, 50)
    }
  }

  //构建表单
  build(fields: SmartField[], values: any): FormGroup {
    //console.log("[SmartEditor] build", fields, values)
    values = values || {}

    let fs: any = {}
    fields?.forEach(f => {

      switch (f.type) {
        case 'object':
          fs[f.key] = this.build(f.children || [], values[f.key])
          break;
        case "list":
        case "table":
          fs[f.key] = this.fb.array(values[f.key]?.map((v: any) => this.build(f.children || [], v)) || [])
          break;
        default:
          if (f.array)
            fs[f.key] = this.fb.array(values[f.key]?.map((v: any) => createControl(f, v)) || [])
          else
            fs[f.key] = createControl(f, values[f.key])
          break;
      }

      //订阅变化
      if (f.change)
        fs[f.key].valueChanges.subscribe((res: any) => f.change?.(res))
      if (f.change_action)
        fs[f.key].valueChanges.subscribe((res: any) => this.action.emit(f.change_action))
    })
    return this.fb.group(fs)
  }

  //设置数据
  setValue(value: any) {
    //console.log("[SmartEditor] setValue", value)
    //this.group.setValue(value)
    this._values = value
    if (this._fields && this._fields.length) {
      this.group = this.build(this._fields, value)
      this.group.valueChanges.subscribe(res => this.change.emit(res))
    }
  }

  //补充数据
  patchValue(value: any) {
    //console.log("[SmartEditor] patchValue", value)
    //setTimeout(() => this.group.patchValue(value))
    //TODO 数组类型 需要创建control
    this.group.patchValue(value)
  }

  get valid() {
    return this.group.valid
  }

  get value() {
    return this.group.value
  }

  getRawValue() {
    //this.group.updateValueAndValidity()
    return this.group.getRawValue()
  }

  constructor(private fb: FormBuilder) {
  }

  ngOnInit(): void {
    if (this._fields && this._fields.length) {
      this.group = this.build(this._fields, this._values)
      this.group.valueChanges.subscribe(res => this.change.emit(res))
    }
  }


  handleUpload(control: FormControl, $event: NzUploadChangeParam) {
    if ($event.type == 'success') {
      //this.group.patchValue({[key]: $event.file.response.data[0]})
      control.setValue($event.file.response.data[0])
    }
  }

  handleUploadImages(control: FormControl, $event: NzUploadChangeParam) {
    let paths: any = []
    $event.fileList.forEach(file => {
      if (file.response?.data?.[0])
        paths.push(file.response.data[0])
    })
    //this.group.patchValue({[key]: paths})
    control.setValue(paths)
  }


  tableDrop(array: FormArray, event: CdkDragDrop<any, any>) {
    moveItemInArray(array.controls, event.previousIndex, event.currentIndex);
    array.updateValueAndValidity()
  }

  tableCopy(fields: SmartField[], array: FormArray, i: number) {
    let control = array.at(i)
    let group = this.build(fields, control.value)
    array.insert(i, group)
  }

  tableRemove(array: FormArray, i: number) {
    array.removeAt(i);
  }

  tableAdd(array: FormArray, fields: SmartField[]) {
    array.push(this.build(fields, {}))
  }

  protected readonly Infinity = Infinity;


  arrayAdd(array: FormArray, field: SmartField) {
    array.push(createControl(field))
  }

  arrayDrop(array: FormArray, event: CdkDragDrop<any, any>) {
    moveItemInArray(array.controls, event.previousIndex, event.currentIndex);
    array.updateValueAndValidity()
  }

  arrayCopy(field: any, array: FormArray, i: number) {
    let control = array.at(i)
    array.insert(i, createControl(field, control.value))
  }

  arrayRemove(array: FormArray, i: number) {
    array.removeAt(i);
  }

  calc_condition(group: FormGroup, condition: SmartFieldCondition) {
    let value: any = group.get(condition.key)?.value
    //console.log('calc_condition', condition, value)

    switch (condition.type) {
      case "=":
      case "==":
        return value == condition.value
      case "!=":
      case "~=":
      case "<>":
        return value != condition.value
      case "<":
        return value < condition.value
      case "<=":
        return value <= condition.value
      case ">=":
        return value >= condition.value
      case ">":
        return value > condition.value
      case "in":
        for (let i=0; i<condition.value.length; i++) {
          if (value == condition.value[i]) {
            return true
          }
        }
        return false
      case "out":
        for (let i=0; i<condition.value.length; i++) {
          if (value == condition.value[i]) {
            return false
          }
        }
        return true
    }


    return false;
  }
}
