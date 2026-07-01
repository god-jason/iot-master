import {Component, ViewChild} from '@angular/core';
import {SmartInfoComponent} from '../../lib/smart-info/smart-info.component';
import {NzSkeletonComponent} from 'ng-zorro-antd/skeleton';
import {SmartToolbarComponent} from '../../lib/smart-toolbar/smart-toolbar.component';

import {TemplateBase} from '../template-base.component';
import {CommonModule} from '@angular/common';


@Component({
  selector: 'app-detail',
  imports: [
    CommonModule,
    SmartInfoComponent,
    NzSkeletonComponent,
    SmartToolbarComponent
],
  templateUrl: './detail.component.html',
  standalone: true,
  styleUrl: './detail.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class DetailComponent extends TemplateBase {
  @ViewChild("toolbar", {static: false}) toolbar!: SmartToolbarComponent;
  toolbarValue = {}


}