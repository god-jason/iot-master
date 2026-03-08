import {Component, ViewChild} from '@angular/core';
import {TemplateBase} from '../template-base.component';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzSpinComponent} from 'ng-zorro-antd/spin';
import {DecimalPipe, NgIf} from '@angular/common';
import {SmartToolbarComponent} from '../../lib/smart-toolbar/smart-toolbar.component';
import {NzColDirective, NzRowDirective} from 'ng-zorro-antd/grid';
import {NzStatisticComponent} from 'ng-zorro-antd/statistic';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {NzIconDirective} from 'ng-zorro-antd/icon';

@Component({
  selector: 'app-statistic',
  imports: [
    NzCardComponent,
    NzSpinComponent,
    NgIf,
    SmartToolbarComponent,
    NzRowDirective,
    NzColDirective,
    NzStatisticComponent,
    DecimalPipe,
    NzButtonComponent,
    NzIconDirective
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
