import { Component, forwardRef, Input } from '@angular/core';
import {
  ControlValueAccessor,
  FormsModule,
  NG_VALUE_ACCESSOR,
  ReactiveFormsModule
} from '@angular/forms';
import { NzTimePickerComponent } from 'ng-zorro-antd/time-picker';
import { NzDatePickerComponent } from 'ng-zorro-antd/date-picker';
import dayjs, { Dayjs } from 'dayjs';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'smart-date-picker',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    NzDatePickerComponent,
    NzTimePickerComponent
  ],
  templateUrl: './smart-date-picker.html',
  styleUrl: './smart-date-picker.scss',
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => SmartDatePicker),
      multi: true
    }
  ]
})
export class SmartDatePicker implements ControlValueAccessor {
  @Input() mode: 'datetime' | 'date' | 'time' = 'datetime';
  @Input() placeholder = '请选择';
  @Input() allowClear = true;

  formats = {
    datetime: 'YYYY-MM-DD HH:mm:ss',
    date: 'YYYY-MM-DD',
    time: 'HH:mm:ss'
  };

  value: Date | null = null;
  disabled = false;

  // ===== 写入值（外部 -> 内部）=====
  writeValue(obj: any): void {
    if (!obj) {
      this.value = null;
      return;
    }

    try {
      if (this.mode === 'time') {
        const d = dayjs(`2000-01-01 ${obj}`);
        this.value = d.isValid() ? d.toDate() : null;
      } else {
        const d = dayjs(obj);
        this.value = d.isValid() ? d.toDate() : null;
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
  onModelChange(value: Date | null) {
    if (!value) {
      this.onChange(null);
      this.onTouched();
      return;
    }

    const fmt = this.formats[this.mode];
    const result = dayjs(value).format(fmt);

    this.onChange(result);
    this.onTouched();
  }
}
