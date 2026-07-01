import {Component, ViewChild} from '@angular/core';
import {TemplateBase} from '../template-base.component';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzSkeletonComponent} from 'ng-zorro-antd/skeleton';
import {CommonModule, DecimalPipe} from '@angular/common';
import {SmartToolbarComponent} from '../../lib/smart-toolbar/smart-toolbar.component';
import {NzColDirective, NzRowDirective} from 'ng-zorro-antd/grid';
import {NzStatisticComponent} from 'ng-zorro-antd/statistic';

@Component({
  selector: 'app-statistic',
  imports: [
    CommonModule,
    NzCardComponent,
    NzSkeletonComponent,
    SmartToolbarComponent,
    NzRowDirective,
    NzColDirective,
    NzStatisticComponent,
    DecimalPipe
  ],
  templateUrl: './statistic.component.html',
  standalone: true,
  styleUrl: './statistic.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class StatisticComponent extends TemplateBase {
  @ViewChild("toolbar", {static: false}) toolbar!: SmartToolbarComponent;
  toolbarValue = {}

}
