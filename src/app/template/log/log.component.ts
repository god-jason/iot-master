import {Component, ElementRef, ViewChild} from '@angular/core';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzSpinComponent} from 'ng-zorro-antd/spin';
import {TemplateBase} from '../template-base.component';
import dayjs from 'dayjs';


@Component({
  selector: 'app-log',
  standalone: true,
  imports: [
    NzCardComponent,
    NzSpinComponent,
  ],
  templateUrl: './log.component.html',
  styleUrl: './log.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class LogComponent extends TemplateBase {
  @ViewChild('box') box!: ElementRef;

  push(text: string) {
    const p = document.createElement('p');
    p.innerText = dayjs().format('YYYY-MM-DD HH:mm:ss') + ' -> ' + text;
    this.box.nativeElement.appendChild(p);
  }

  insert(text: string) {
    const p = document.createElement('p');
    p.innerText = dayjs().format('YYYY-MM-DD HH:mm:ss') + ' -> ' + text;
    this.box.nativeElement.prepend(p);
  }

}
