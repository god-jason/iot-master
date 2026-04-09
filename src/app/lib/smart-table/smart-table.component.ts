import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NzIconDirective} from "ng-zorro-antd/icon";
import {Data, Router} from "@angular/router";
import {NzTableFilterList, NzTableModule, NzTableQueryParams} from "ng-zorro-antd/table";
import {CommonModule} from "@angular/common";
import {NzPopconfirmDirective} from "ng-zorro-antd/popconfirm";
import {FormsModule} from '@angular/forms';
import {NzModalModule, NzModalService} from 'ng-zorro-antd/modal';
import {SmartRequestService} from '../smart-request.service';
import {NzBytesPipe} from 'ng-zorro-antd/pipes';
import {NzProgressComponent} from 'ng-zorro-antd/progress';
import {NzTagComponent} from 'ng-zorro-antd/tag';
import {NzAvatarComponent} from 'ng-zorro-antd/avatar';


export interface SmartAction {
  type: 'link' | 'script' | 'page' | 'dialog'
  link?: string
  link_func?: string | Function | ((data: any, index: number) => string)
  params?: any
  params_func?: string | Function | ((data: any, index: number) => any)
  script?: string | Function | ((data: any, index: number) => string)
  after_close?: string | Function | ((result: any, data: any, index: number) => string) //dialog回调
  app?: string
  page?: string
  dialog?: boolean
  external?: boolean
}

export interface SmartActionRow {
  action: SmartAction
  data: any
  index: number
}

export interface SmartTableColumn {
  key: string
  label: string
  type?: string
  format?: string
  keyword?: boolean
  sortable?: boolean
  filter?: NzTableFilterList
  date?: boolean
  ellipsis?: boolean
  break?: boolean
  action?: SmartAction

  //仅限管理员
  admin?: boolean
  //仅限非管理员
  not_admin?: boolean
}

export interface SmartTableOperator {
  icon?: string
  label?: string
  title?: string
  action: SmartAction
  confirm?: string

  //仅限管理员
  admin?: boolean
  //仅限非管理员
  not_admin?: boolean
}

export interface SmartTableButton {
  icon?: string
  label: string
  title?: string
  action?: SmartAction
}


export interface SmartTableParams {
  buttons?: SmartTableButton[];
  columns: SmartTableColumn[]
  operators: SmartTableOperator[];
}


export interface ParamJoin {
  table: string
  local_field: string
  foreign_field: string
  field: string
  as: string
}

export interface ParamSearch {
  filter: { [key: string]: any }
  skip?: number
  limit?: number
  sort?: { [key: string]: number }
  keyword?: { [key: string]: string }
  fields?: string[]
  joins?: ParamJoin[]
}

@Component({
  selector: 'smart-table',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    NzTableModule,
    NzPopconfirmDirective,
    NzModalModule,
    NzIconDirective,
    NzBytesPipe,
    NzProgressComponent,
    NzTagComponent,
    NzAvatarComponent,
  ],
  templateUrl: './smart-table.component.html',
  styleUrl: './smart-table.component.scss'
})
export class SmartTableComponent implements OnInit {

  @Input() pageSize = parseInt(localStorage.getItem("table_page_size") || "10"); //默认10页
  @Input() pageIndex = 1;

  _columns: SmartTableColumn[] = []
  @Input() set columns(cols: SmartTableColumn[]){
    this._columns = cols?.filter(f=>{
      //管理员
      if (f.admin)
        return this.user?.admin
      //非管理员
      if (f.not_admin)
        return !(this.user?.admin)
      return true
    })
  }
  get columns(){
    return this._columns
  }

  _operators: SmartTableOperator[] = []
  @Input() set operators(ops: SmartTableOperator[]){
    this._operators = ops?.filter(f=>{
      //管理员
      if (f.admin)
        return this.user?.admin
      //非管理员
      if (f.not_admin)
        return !(this.user?.admin)
      return true
    })
  }
  get operators(){
    return this._operators
  }


  @Input() datum: any[] = [];
  @Input() total: number = 0;
  @Input() loading = false;

  //@Input() showSearch: boolean = true

  @Output() query = new EventEmitter<ParamSearch>
  @Output() action = new EventEmitter<SmartActionRow>();
  @Input() user:any = {}

  //多选
  @Input() batch = false;
  checked = false;
  indeterminate = false;
  selects = new Set<number>();

  body: ParamSearch = {filter: {}}

  constructor(private router: Router, private ms: NzModalService, private request: SmartRequestService) {
  }


  updateCheckedSet(id: number, checked: boolean): void {
    if (checked) {
      this.selects.add(id);
    } else {
      this.selects.delete(id);
    }
  }

  onCurrentPageDataChange(listOfCurrentPageData: readonly Data[]): void {
    this.refreshCheckedStatus();
  }

  refreshCheckedStatus(): void {
    this.checked = this.datum.every(({id}) => this.selects.has(id));
    this.indeterminate = this.datum.some(({id}) => this.selects.has(id)) && !this.checked;
  }

  onItemChecked(id: number, checked: boolean): void {
    this.updateCheckedSet(id, checked);
    this.refreshCheckedStatus();
  }

  onAllChecked(checked: boolean): void {
    this.datum.forEach(({id}) => this.updateCheckedSet(id, checked));
    this.refreshCheckedStatus();
  }


  ngOnInit(): void {
  }

  onQuery(query: NzTableQueryParams) {
    //console.log("table view onQuery", query)
    //过滤器
    query.filter.forEach(f => {
      if (f.value) {
        if (f.value.length > 1)
          this.body.filter[f.key] = f.value;
        else if (f.value.length === 1)
          this.body.filter[f.key] = f.value[0];
      }
    })

    //分页
    this.body.skip = (query.pageIndex - 1) * query.pageSize;
    this.body.limit = query.pageSize;

    //排序
    const sorts = query.sort.filter(s => s.value);
    if (sorts.length) {
      this.body.sort = {};
      sorts.forEach(s => {
        // @ts-ignore
        this.body.sort[s.key] = s.value === 'ascend' ? 1 : -1;
      });
    } else {
      delete this.body.sort;
    }

    this.query.emit(this.body)
  }

  protected readonly parseInt = parseInt;

  protected onPageSizeChange($event: number) {
    localStorage.setItem("table_page_size", $event + '')
  }
}
