import {Component} from '@angular/core';
import {RouterLink} from "@angular/router";
import {NzButtonComponent} from "ng-zorro-antd/button";
import {NzResultComponent} from 'ng-zorro-antd/result';

@Component({
  selector: 'app-unknown',
  standalone: true,
  imports: [
    RouterLink,
    NzButtonComponent,
    NzResultComponent,
  ],
  templateUrl: './unknown.component.html',
  styleUrl: './unknown.component.scss'
})
export class UnknownComponent {

}
