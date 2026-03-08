import {Component, ViewChild} from '@angular/core';
import {SmartInfoComponent} from '../../lib/smart-info/smart-info.component';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {NzSpinComponent} from 'ng-zorro-antd/spin';
import {NzIconDirective} from 'ng-zorro-antd/icon';
import {SmartToolbarComponent} from '../../lib/smart-toolbar/smart-toolbar.component';
import {NgIf} from '@angular/common';
import {TemplateBase} from '../template-base.component';


@Component({
  selector: 'app-info',
  imports: [
    SmartInfoComponent,
    NzCardComponent,
    NzButtonComponent,
    NzSpinComponent,
    NzIconDirective,
    SmartToolbarComponent,
    NgIf
  ],
  templateUrl: './info.component.html',
  standalone: true,
  styleUrl: './info.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class InfoComponent extends TemplateBase {
  @ViewChild("toolbar", {static: false}) toolbar!: SmartToolbarComponent;
  toolbarValue = {}


}
