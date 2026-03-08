import {Component} from '@angular/core';
import {MarkdownComponent as md} from 'ngx-markdown';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzSpinComponent} from 'ng-zorro-antd/spin';
import {TemplateBase} from '../template-base.component';


@Component({
  selector: 'app-markdown',
  standalone: true,
  imports: [
    md,
    NzCardComponent,
    NzSpinComponent,
  ],
  templateUrl: './markdown.component.html',
  styleUrl: './markdown.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class MarkdownComponent extends TemplateBase {

}
