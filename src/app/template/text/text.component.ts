import {Component} from '@angular/core';
import {TemplateBase} from '../template-base.component';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzSpinComponent} from 'ng-zorro-antd/spin';
import {CommonModule} from '@angular/common';


@Component({
  selector: 'app-text',
  standalone: true,
  imports: [
    CommonModule,
    NzCardComponent,
    NzSpinComponent,
  ],
  templateUrl: './text.component.html',
  styleUrl: './text.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class TextComponent extends TemplateBase {

}
