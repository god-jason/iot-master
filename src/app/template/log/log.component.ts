import {Component, ElementRef, ViewChild} from '@angular/core';
import {NzSkeletonComponent} from 'ng-zorro-antd/skeleton';
import {TemplateBase} from '../template-base.component';
import dayjs from 'dayjs';
import {CommonModule} from '@angular/common';


@Component({
  selector: 'app-log',
  standalone: true,
  imports: [
    CommonModule,
    NzSkeletonComponent,
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
