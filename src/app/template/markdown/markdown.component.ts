import {Component} from '@angular/core';
import {MarkdownComponent as md} from 'ngx-markdown';
import {NzSkeletonComponent} from 'ng-zorro-antd/skeleton';
import {TemplateBase} from '../template-base.component';
import {CommonModule} from '@angular/common';


@Component({
  selector: 'app-markdown',
  standalone: true,
  imports: [
    CommonModule,
    md,
    NzSkeletonComponent,
  ],
  templateUrl: './markdown.component.html',
  styleUrl: './markdown.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class MarkdownComponent extends TemplateBase {

}
