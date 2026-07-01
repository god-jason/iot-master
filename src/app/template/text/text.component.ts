import {Component} from '@angular/core';
import {TemplateBase} from '../template-base.component';
import {NzSkeletonComponent} from 'ng-zorro-antd/skeleton';
import {CommonModule} from '@angular/common';


@Component({
  selector: 'app-text',
  standalone: true,
  imports: [
    CommonModule,
    NzSkeletonComponent,
  ],
  templateUrl: './text.component.html',
  styleUrl: './text.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class TextComponent extends TemplateBase {

}
