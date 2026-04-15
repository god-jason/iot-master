import {Component, forwardRef, Input} from '@angular/core';
import {ControlValueAccessor, FormsModule, NG_VALUE_ACCESSOR, ReactiveFormsModule} from '@angular/forms';
import {NzTimePickerComponent} from 'ng-zorro-antd/time-picker';
import {NzDatePickerComponent} from 'ng-zorro-antd/date-picker';
import dayjs from 'dayjs';
import {CommonModule} from '@angular/common';

@Component({
  selector: 'smart-date-picker',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    NzDatePickerComponent,
    NzTimePickerComponent,
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
  @Input() mode = "datetime"

  formats: any = {
    datetime: 'YYYY-MM-DD HH:mm:ss',
    date: 'YYYY-MM-DD',
    time: 'HH:mm:ss',
  }

  value: Date = new Date()
  disabled = false

  writeValue(obj: any): void {
    if (obj) {
      if (this.mode == "time")
        this.value = dayjs("2000-01-01 " + obj).toDate();
      else
        this.value = dayjs(obj).toDate();
    }
  }

  onChange = (val: string) => {
  };
  onTouched = () => {
  };

  registerOnChange(fn: any): void {
    this.onChange = fn
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn
  }

  setDisabledState?(isDisabled: boolean): void {
    this.disabled = isDisabled
  }

  onModelChange($event: any) {
    let fmt = this.formats[this.mode]
    this.onChange(dayjs($event).format(fmt))
    this.onTouched()
  }

}
