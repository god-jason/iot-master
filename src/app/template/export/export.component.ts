import { Component } from '@angular/core';
import {TemplateBase} from '../template-base.component';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {NzIconDirective} from 'ng-zorro-antd/icon';
import {NzSpinComponent} from 'ng-zorro-antd/spin';

@Component({
  selector: 'app-export',
  imports: [
    NzCardComponent,
    NzButtonComponent,
    NzIconDirective,
    NzSpinComponent
  ],
  templateUrl: './export.component.html',
  standalone: true,
  styleUrl: './export.component.scss'
})
export class ExportComponent extends TemplateBase {

  exporting = false

  export() {

  }

}
