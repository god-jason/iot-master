import {GuardResult, MaybeAsync, Router, Routes, UrlSegment} from '@angular/router';
import {LoginComponent} from './login/login.component';
import {UserService} from './user.service';
import {inject} from '@angular/core';
import {UnknownComponent} from './unknown/unknown.component';
import {PageComponent} from './page/page.component';
import {SettingComponent} from './setting/setting.component';
import {AdminComponent} from './admin/admin.component';
import {PasswordComponent} from './password/password';

export const adminRoutes: Routes = [
  //{path: '', pathMatch: 'full', redirectTo: ''},
  {path: 'login', component: LoginComponent},
  {path: 'password', component: PasswordComponent},
  {
    path: '',
    canActivate: [loginGuard],
    component: AdminComponent,
    children: [
      {path: '', pathMatch: 'full', redirectTo: 'page/dash'},
      {path: 'page/:page', component: PageComponent},
      //{path: 'page/:app/:page', component: PageComponent},
      {path: 'setting/:module', component: SettingComponent},
      {path: '**', component: UnknownComponent},
    ],
    //子模块
    //loadChildren: () => import('./admin/admin.module').then(m => m.AdminModule),
  },
  //{path: 'app/:app/page/:page', component: PageComponent},
  {path: '**', component: UnknownComponent},
];

function loginGuard(router: Router, segments: UrlSegment[]): MaybeAsync<GuardResult> {
  let us = inject(UserService)
  if (us.valid()) return true
  else return router.parseUrl("/login");
}
