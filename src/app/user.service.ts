import {Injectable} from '@angular/core';
import {SmartRequestService} from './lib/smart-request.service';
import {Router} from '@angular/router';

@Injectable({
  providedIn: 'root'
})
export class UserService {

  private _user: any

  constructor(protected request: SmartRequestService, protected router: Router) {
    //TODO 自动加载登录状态，此处应该有token
    let u = localStorage.getItem('user')
    if (u) {
      this._user = JSON.parse(u)
    } else {
      request.get('/api/me').subscribe((res) => {
        console.log("api me", res)
        if (res.error) {
          this.router.navigateByUrl("/login")
          return
        }
        this.set(res.data);
      });
    }
  }

  get user() {
    return this._user || {};
  }

  set(user: any) {
    this._user = user;
    localStorage.setItem("user", JSON.stringify(user));
  }

  unset(user: any) {
    this._user = undefined;
    localStorage.removeItem("user");
  }

  valid() {
    return !!this._user
  }


}
