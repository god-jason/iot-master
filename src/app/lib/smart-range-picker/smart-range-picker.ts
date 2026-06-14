import { Component, forwardRef, Input } from '@angular/core';
import {
  ControlValueAccessor,
  FormsModule,
  NG_VALUE_ACCESSOR,
  ReactiveFormsModule
} from '@angular/forms';
import { NzDatePickerModule } from 'ng-zorro-antd/date-picker';
import dayjs from 'dayjs';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'smart-range-picker',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    NzDatePickerModule
  ],
  templateUrl: './smart-range-picker.html',
  styleUrl: './smart-range-picker.scss',
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => SmartRangePicker),
      multi: true
    }
  ]
})
export class SmartRangePicker implements ControlValueAccessor {
  @Input() mode: 'daterange' | 'datetimerange' = 'daterange';
  @Input() placeholder = ['开始日期', '结束日期'];
  @Input() allowClear = true;

  formats = {
    daterange: 'YYYY-MM-DD',
    datetimerange: 'YYYY-MM-DD HH:mm:ss'
  };

  // 值格式: [Date | null, Date | null]
  value: [Date | null, Date | null] | null = null;
  disabled = false;

  // ===== 写入值（外部 -> 内部）=====
  writeValue(obj: any): void {
    if (!obj) {
      this.value = null;
      return;
    }

    try {
      // 支持字符串数组格式 ["2024-01-01 00:00:00", "2024-01-31 23:59:59"]
      if (Array.isArray(obj) && obj.length === 2) {
        const start = obj[0] ? dayjs(obj[0]) : null;
        const end = obj[1] ? dayjs(obj[1]) : null;
        this.value = start && end ? [start.toDate(), end.toDate()] : null;
      } else {
        this.value = null;
      }
    } catch {
      this.value = null;
    }
  }

  // ===== 注册 =====
  onChange: (val: any) => void = () => {};
  onTouched: () => void = () => {};

  registerOnChange(fn: any): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
  }

  // ===== 内部变化（内部 -> 外部）=====
  onModelChange(value: [Date | null, Date | null] | null) {
    if (!value || !value[0] || !value[1]) {
      this.onChange(null);
      this.onTouched();
      return;
    }

    const fmt = this.formats[this.mode];
    const result: [string, string] = [
      dayjs(value[0]).format(fmt),
      dayjs(value[1]).format(fmt)
    ];

    this.onChange(result);
    this.onTouched();
  }
}
