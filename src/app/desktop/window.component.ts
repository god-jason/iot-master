import {Component, ElementRef, EventEmitter, inject, Input, OnInit, Output, ViewChild,} from '@angular/core';
import {NzMessageService} from 'ng-zorro-antd/message';
import {DomSanitizer, SafeResourceUrl} from '@angular/platform-browser';
import {NgForOf, NgIf, NgStyle} from '@angular/common';
import {CdkDrag, CdkDragHandle} from '@angular/cdk/drag-drop';
import {NzSpaceComponent, NzSpaceItemDirective} from 'ng-zorro-antd/space';
import {NzButtonComponent} from 'ng-zorro-antd/button';
import {NzIconDirective} from 'ng-zorro-antd/icon';
import {NzTabComponent, NzTabSetComponent} from 'ng-zorro-antd/tabs';
import {FullscreenDirective} from './fullscreen.directive';


export interface WindowDialog {
  show: boolean
  url: string
  title: string
  zIndex: number
  id: number
}

@Component({
  selector: 'app-window',
  templateUrl: './window.component.html',
  styleUrls: ['./window.component.scss'],
  imports: [
    NgStyle,
    CdkDrag,
    CdkDragHandle,
    NzSpaceComponent,
    NzSpaceItemDirective,
    NzButtonComponent,
    NzIconDirective,
    NzTabSetComponent,
    NgIf,
    NzTabComponent,
    NgForOf,
    FullscreenDirective
  ],
  standalone: true
})
export class WindowComponent implements OnInit {
  san = inject(DomSanitizer)

  // index = 0;
  @Input() id: number = 0;
  @Input() index: number = 0;
  @Input() title: string = '';
  @Input() show: boolean = false;

  _url!: SafeResourceUrl;
  @Input() set url(u: string) {
    this._url = this.san.bypassSecurityTrustResourceUrl(u)
  }

  @Output() onClose = new EventEmitter();
  @Output() onHide = new EventEmitter();
  @Output() onIndexChange = new EventEmitter<number>();


  width: any = '60vw';
  height: any = '50vh';

  dragPosition = {x: 0, y: 0};

  dynamic = false;
  items: any[] = [];

  ngOnInit(): void {
  }

  tabData: any;

  setIndex() {
    this.onIndexChange.emit(this.id);
  }

  close() {
    this.onClose.emit(this.id);
    this.width = '60vw';
    this.height = '50vh';
  }

  minimize() {
    this.onHide.emit(this.id);
  }

  showTab() {
  }

  maximize() {
    this.dynamic = !this.dynamic;
    this.dragPosition = {x: 0, y: 0};

    if (this.dynamic) {
      this.width = '100vw';
      this.height = '100vh';
    } else {
      this.width = '60vw';
      this.height = '50vh';
    }
  }

  @ViewChild('iframe') iframe!: ElementRef;

  back(){
    //console.log(this.iframe.nativeElement)
    this.iframe.nativeElement.contentWindow.history.back()
  }

  forward(){
    this.iframe.nativeElement.contentWindow.history.forward()
  }
}
