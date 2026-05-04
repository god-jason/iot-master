import {Component} from '@angular/core';
import {TemplateBase} from '../template-base.component';


@Component({
  selector: 'app-blank',
  imports: [
],
  templateUrl: './blank.component.html',
  standalone: true,
  styleUrl: './blank.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class BlankComponent extends TemplateBase {


}
