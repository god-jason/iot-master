import {Component} from '@angular/core';
import {Router, RouterLink} from '@angular/router';
import {NzModalModule, NzModalService} from 'ng-zorro-antd/modal';
import {UserService} from '../user.service';
import {NzMessageService} from 'ng-zorro-antd/message';
import {SmartRequestService} from '../lib/smart-request.service';
import {NzContentComponent, NzFooterComponent, NzHeaderComponent, NzLayoutComponent} from 'ng-zorro-antd/layout';
import {NzColDirective, NzRowDirective} from 'ng-zorro-antd/grid';
import {NgClass, NgForOf, NgIf} from '@angular/common';
import {NzIconDirective} from 'ng-zorro-antd/icon';
import {WindowComponent, WindowDialog} from './window.component';
import {NzDropDownDirective, NzDropdownMenuComponent} from 'ng-zorro-antd/dropdown';
import {NzMenuDirective, NzMenuDividerDirective, NzMenuItemComponent, NzSubMenuComponent} from 'ng-zorro-antd/menu';
import {NzConfigService} from 'ng-zorro-antd/core/config';
import {ThemeService} from '../theme.service';

@Component({
  selector: 'app-desktop',
  templateUrl: './desktop.component.html',
  styleUrls: ['./desktop.component.scss'],
  imports: [
    NzLayoutComponent,
    NzHeaderComponent,
    NzContentComponent,
    NzRowDirective,
    NzColDirective,
    NgForOf,
    NzFooterComponent,
    NzIconDirective,
    WindowComponent,
    NgIf,
    NzModalModule,
    NzDropDownDirective,
    NzDropdownMenuComponent,
    NzMenuDirective,
    NzMenuItemComponent,
    NzSubMenuComponent,
    RouterLink,
    NgClass,
    NzMenuDividerDirective,
  ],
  standalone: true
})
export class DesktopComponent {
  title: any;
  show: any;

  apps: any[] = []

  oem: any = {
    name: 'BOAT',
    logo: '/boat.svg',
    company: '南京本易物联网有限公司',
  }

  version: any = {
    version: '',
    build: '',
    git: '',
  }

  menus: any[] = []
  settings: any[] = []
  primaryColor: any


  windows: WindowDialog[] = [];


  userInfo: any;
  showMenu = false;
  appIndex: any = {};

  constructor(
    private router: Router,
    private rs: SmartRequestService,
    private nzConfigService: NzConfigService,
    protected ts: ThemeService,
    private ms: NzModalService,
    private us: UserService,
    private msg: NzMessageService
  ) {
    this.loadApps()
    this.loadOem()
    this.loadMenu()
    this.loadSetting()


    //主题色
    this.primaryColor = localStorage.getItem("primaryColor")
    if (this.primaryColor)
      this.nzConfigService.set('theme', {primaryColor: this.primaryColor})
  }

  loadApps() {
    this.rs.get("shortcuts").subscribe((res) => {
      if (res.error) return
      this.apps = res.data;
    })
  }


  loadVersion() {
    this.rs.get("version").subscribe((res) => {
      if (res.error) return
      Object.assign(this.version, res.data);
    })
  }

  loadOem() {
    this.rs.get("oem").subscribe((res) => {
      if (res.error) return
      Object.assign(this.oem, res.data);
    })
  }

  loadMenu() {
    this.rs.get("menus").subscribe((res) => {
      if (res.error) return
      this.menus = res.data
    })
  }

  loadSetting() {
    this.rs.get("settings").subscribe((res) => {
      if (res.error) return
      this.settings = res.data
    })
  }

  onHide(id: any) {
    this.windows.filter((item: any, index: any) => {
      if (item.id === id) {
        item.show = false;
        item.tab = true;
      }
    });
  }

  onClose(id: any) {
    this.windows.filter((item: any, index: any) => {
      if (item.id === id) {
        this.windows.splice(index, 1);
      }
    });
  }

  activeTab(id: number) {
    this.windows.filter((item: WindowDialog, index: any) => {
      if (item.id === id) {
        item.show = true;
        //item.tab = false;
      }
    });
    this.activeWindow(id);
  }

  activeWindow(id: number) {
    this.windows.filter((item: WindowDialog, index: any) => {
      item.zIndex = 0;
      if (item.id === id) {
        item.zIndex = 9999;
      }
    });
  }


  idIncrement = 0

  openMenu(app: any) {
    console.log("open", app)
    if (window.innerWidth < 800) {
      this.router.navigate([app.entries[0].path]);
      return;
    }

    let win: WindowDialog = {
      show: true,
      url: app.url,
      title: app.name,
      zIndex: 0,
      id: this.idIncrement++,
    }

    this.windows.push(win);

    this.activeTab(win.id);
    this.activeWindow(win.id);
  }

  setMenu(status: any, name: any) {
    //this._as.apps[this.appIndex[name]].status = !status;
    this.msg.success('设置成功');
  }

  logout() {
    localStorage.removeItem("token")
    this.rs.get('logout').subscribe((res) => {
    }).add(() => this.router.navigateByUrl('/login'));
  }

  switchAdmin() {
    localStorage.setItem("ui-mode", "admin")
    window.location.href = "/"
  }
}
