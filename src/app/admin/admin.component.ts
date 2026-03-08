import {Component} from '@angular/core';
import {NzContentComponent, NzHeaderComponent, NzLayoutComponent, NzSiderComponent} from 'ng-zorro-antd/layout';
import {NzDropDownDirective, NzDropdownMenuComponent} from 'ng-zorro-antd/dropdown';
import {NzIconDirective} from 'ng-zorro-antd/icon';
import {NzMenuDirective, NzMenuDividerDirective, NzMenuItemComponent, NzSubMenuComponent} from 'ng-zorro-antd/menu';
import {Router, RouterLink, RouterOutlet} from '@angular/router';
import {UserService} from '../user.service';
import {SmartRequestService} from '../lib/smart-request.service';
import {NzConfigService} from 'ng-zorro-antd/core/config';
import {Location} from '@angular/common';
import {ThemeService} from '../theme.service';
import {NzButtonComponent} from 'ng-zorro-antd/button';

@Component({
  selector: 'app-admin',
  standalone: true,
  imports: [
    NzContentComponent,
    NzDropDownDirective,
    NzDropdownMenuComponent,
    NzHeaderComponent,
    NzIconDirective,
    NzLayoutComponent,
    NzMenuDirective,
    NzMenuDividerDirective,
    NzMenuItemComponent,
    NzSiderComponent,
    NzSubMenuComponent,
    RouterLink,
    RouterOutlet,
    NzButtonComponent,
  ],
  templateUrl: './admin.component.html',
  styleUrl: './admin.component.scss'
})
export class AdminComponent {
  isCollapsed = false;

  oem: any = {
    name: 'BOAT',
    logo: '/boat.svg',
    company: '南京本易物联网有限公司',
  }
  version: any = {
    version: '',
    build: '',
    git: '',
    runtime: '',
  }

  colors = [
    {value:'#188ffe',name:'商务蓝'},
    {value:'#712ed0',name:'高贵紫'},
    {value:'#13c0c1',name:'蓝绿色'},
    {value:'#51c21b',name:'清新绿'},
    {value:'#e92f96',name:'洋红色'},
    {value:'#f3222e',name:'红色'},
    {value:'#f88b17',name:'橙色'},
    {value:'#f8d915',name:'黄色'},
    {value:'#f8531d',name:'绯红色'},
    {value:'#2f53ea',name:'极客蓝'},
    {value:'#9fd712',name:'绿黄色'},
    {value:'#f8ac15',name:'土豪金'},
  ];

  menus: any[] = []
  settings: any[] = []
  primaryColor: any

  constructor(protected us: UserService,
              private request: SmartRequestService,
              private nzConfigService: NzConfigService,
              private router: Router,
              protected ts: ThemeService,
              protected location: Location,
  ) {
    this.loadOem()
    this.loadMenu()
    this.loadSetting()
    this.loadVersion()

    //主题色
    this.primaryColor = localStorage.getItem("primaryColor")
    if (this.primaryColor)
      this.nzConfigService.set('theme', {primaryColor: this.primaryColor})
  }

  loadOem() {
    this.request.get("oem").subscribe((res) => {
      if (res.error) return
      Object.assign(this.oem, res.data);
    })
  }

  loadVersion() {
    this.request.get("version").subscribe((res) => {
      if (res.error) return
      Object.assign(this.version, res.data);
    })
  }

  loadMenu() {
    this.request.get("menus").subscribe((res) => {
      if (res.error) return
      this.menus = res.data
    })
  }

  loadSetting() {
    this.request.get("settings").subscribe((res) => {
      if (res.error) return
      this.settings = res.data
    })
  }

  onChangePrimaryColor(color: any) {
    //主题色
    this.primaryColor = color
    localStorage.setItem("primaryColor", this.primaryColor)
    this.nzConfigService.set('theme', {primaryColor: this.primaryColor})
  }

  onChangeTheme(theme: string) {
    this.ts.changeTheme(theme)
  }

  logout() {
    localStorage.removeItem("token")
    this.request.get("logout").subscribe(res=>{
      this.router.navigateByUrl("/login")
    })
  }

  switchDesktop() {
    localStorage.setItem("ui-mode", "desktop")
    window.location.href = "/"
  }
}
