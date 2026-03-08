import {Inject, Injectable, InjectionToken, Optional} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {catchError, map} from 'rxjs/operators';
import {Observable, of} from 'rxjs';
import {Router} from '@angular/router';
import {NzNotificationService} from "ng-zorro-antd/notification";

export const SMART_API_BASE = new InjectionToken<string>('api_base');

@Injectable({
  providedIn: 'root'
})
export class SmartRequestService {

  public base = '/api/'; //使用ng brokers proxy.config.json
  //public base = environment.host;

  constructor(private ns: NzNotificationService,
              private http: HttpClient,
              private route: Router,
              @Optional() @Inject(SMART_API_BASE) base?: string,
  ) {
    if (base)
      this.base = base
  }

  request(method: string, uri: string, options: any): Observable<any> {
    // 携带Cookie，保持session会话
    options.withCredentials = true;

    //JWT
    let token = localStorage.getItem("token")
    if (token) {
      options.headers = options.headers || {};
      options.headers.Authorization = "Bearer "+token;
    }

    //默认添加基础路径
    if (uri.substring(0, 1) != '/')
      uri = this.base + uri

    return this.http.request<any>(method, uri, options).pipe(
      // 捕捉异常，数据转换
      catchError(err => {
        if (err.status === 404) {
          return of({error: '404 not found: [' + method + '] ' + uri});
        } else if (err.status === 401) {
          // window.location.href = '/login';
          this.route.navigate(['/login']);
          return of({error: '未登录'});
        }
        return of({error: err.message});
      }),
      // 统一错误处理
      map((ret: any) => {
        if (ret && ret.error) {
          if (ret.error === 'Token not found' || ret.error === 'jwt expired') {
            this.route.navigate(['/login']);
          }
          // 有错误统一显示并不是好的做法
          this.ns.error('错误', ret.error);
          throw ret.error; //不抛出Error类型，方便外面直接处理
        }
        return ret;
      })
    );
  }

  get(uri: string, params?: { [k: string]: any }): Observable<any> {
    return this.request('GET', uri, {params});
  }

  put(uri: string, body: any | null, params?: { [k: string]: any }): Observable<any> {
    return this.request('PUT', uri, {params, body});
  }

  post(uri: string, body: any | null, params?: { [k: string]: any }): Observable<any> {
    return this.request('POST', uri, {params, body});
  }

  patch(uri: string, body: any | null, params?: { [k: string]: any }): Observable<any> {
    return this.request('PATCH', uri, {params, body});
  }

  delete(uri: string, params?: { [k: string]: any }): Observable<any> {
    return this.request('DELETE', uri, {params});
  }
}
