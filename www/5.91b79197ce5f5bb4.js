(self.webpackChunkiot_master_ui=self.webpackChunkiot_master_ui||[]).push([[5],{5545:(H,P,a)=>{"use strict";a.d(P,{u:()=>t});var p=a(1764),O=a(4650);let t=(()=>{class T{transform(F,C){return p(F).format(C||"YYYY-MM-DD HH:mm:ss")}}return T.\u0275fac=function(F){return new(F||T)},T.\u0275pipe=O.Yjl({name:"date",type:T,pure:!0}),T})()},2451:(H,P,a)=>{"use strict";a.d(P,{k:()=>y});var p=a(4650),b=a(9651),O=a(6616),t=a(7044),T=a(1664),$=a(7423),F=a(3224);const C=function(){return{authorization:"authorization-text"}};let y=(()=>{class d{constructor(z){this.msg=z,this.onLoad=new p.vpe,this.uploading=!1}load(){this.onLoad.emit()}handleChange(z){if("error"!==z.type){if(z.file&&z.file.response){const U=z.file.response;U.error?this.msg.error(`${U.error}`):(this.msg.success(`\u6210\u529f\u5bfc\u5165${U.data}\u6761\u6570\u636e!`),this.load())}}else this.msg.error("\u4e0a\u4f20\u5931\u8d25")}}return d.\u0275fac=function(z){return new(z||d)(p.Y36(b.dD))},d.\u0275cmp=p.Xpm({type:d,selectors:[["app-import"]],inputs:{url:"url"},outputs:{onLoad:"onLoad"},decls:4,vars:4,consts:[[3,"nzAction","nzHeaders","nzShowUploadList","nzChange"],["nz-button","","nzType","primary"],["nz-icon","","nzType","cloud-upload"]],template:function(z,U){1&z&&(p.TgZ(0,"nz-upload",0),p.NdJ("nzChange",function(D){return U.handleChange(D)}),p.TgZ(1,"button",1),p._UZ(2,"span",2),p._uU(3," \u5bfc\u5165 "),p.qZA()()),2&z&&p.Q6J("nzAction","/api/"+U.url)("nzHeaders",p.DdM(3,C))("nzShowUploadList",!1)},dependencies:[O.ix,t.w,T.dQ,$.Ls,F.FY],styles:[".fileButton[_ngcontent-%COMP%]{height:32px;line-height:32px;border-radius:3px;background-color:#1890ff;border-color:#1890ff;color:#fff;font-weight:400;display:inline-block;padding:0 15px;cursor:pointer}"]}),d})()},297:(H,P,a)=>{"use strict";a.d(P,{B:()=>y});var p=a(4650),b=a(5635),O=a(6616),t=a(7044),T=a(1664),$=a(433);function F(d,m){if(1&d){const z=p.EpF();p.TgZ(0,"button",3),p.NdJ("click",function(){p.CHM(z);const E=p.oxw();return p.KtG(E.onSearch.emit(E.text))}),p._uU(1),p.qZA()}if(2&d){const z=p.oxw();p.xp6(1),p.Oqu(z.searchText)}}const C=function(){return{standalone:!0}};let y=(()=>{class d{constructor(){this.searchText="\u641c\u7d22",this.placeholder="\u5173\u952e\u5b57",this.text="",this.onSearch=new p.vpe}}return d.\u0275fac=function(z){return new(z||d)},d.\u0275cmp=p.Xpm({type:d,selectors:[["app-search-box"]],inputs:{searchText:"searchText",placeholder:"placeholder",text:"text"},outputs:{onSearch:"onSearch"},decls:4,vars:5,consts:[["nzSearch","",3,"nzAddOnAfter"],["type","text","nz-input","",3,"placeholder","ngModel","ngModelOptions","ngModelChange"],["suffixButton",""],["nz-button","","nzType","primary","nzSearch","",3,"click"]],template:function(z,U){if(1&z&&(p.TgZ(0,"nz-input-group",0)(1,"input",1),p.NdJ("ngModelChange",function(D){return U.text=D}),p.qZA()(),p.YNc(2,F,2,1,"ng-template",null,2,p.W1O)),2&z){const E=p.MAs(3);p.Q6J("nzAddOnAfter",E),p.xp6(1),p.Q6J("placeholder",U.placeholder)("ngModel",U.text)("ngModelOptions",p.DdM(4,C))}},dependencies:[b.Zp,b.gB,O.ix,t.w,T.dQ,$.Fj,$.JJ,$.On]}),d})()},5380:(H,P,a)=>{"use strict";function p(b,O){typeof O.filter>"u"&&(O.filter={}),b.filter.forEach(T=>{T.value.length>1?O.filter[T.key]=T.value:1===T.value.length&&(O.filter[T.key]=T.value[0])}),O.skip=(b.pageIndex-1)*b.pageSize,O.limit=b.pageSize;const t=b.sort.filter(T=>T.value);t.length?(O.sort={},t.forEach(T=>{O.sort[T.key]="ascend"===T.value?1:-1})):delete O.sort}a.d(P,{r:()=>p})},3005:(H,P,a)=>{"use strict";a.r(P),a.d(P,{UserModule:()=>Dt});var p=a(6895),b=a(4575),O=a(1346),t=a(4650);let T=(()=>{class e{}return e.\u0275fac=function(n){return new(n||e)},e.\u0275cmp=t.Xpm({type:e,selectors:[["app-me"]],decls:2,vars:0,template:function(n,o){1&n&&(t.TgZ(0,"p"),t._uU(1,"me works!"),t.qZA())}}),e})();var $=a(624),F=a(5380),C=a(235),y=a(433),d=a(1445),m=a(9651),z=a(1243),U=a(8231),E=a(3679),D=a(6704),B=a(5635);function N(e,u){1&e&&(t.ynx(0),t._uU(1,"\u8bf7\u8f93\u5165\u540d\u79f0!"),t.BQk())}function Q(e,u){1&e&&t.YNc(0,N,2,0,"ng-container",17),2&e&&t.Q6J("ngIf",u.$implicit.hasError("required"))}function R(e,u){1&e&&(t.ynx(0),t._uU(1,"\u8bf7\u8f93\u5165\u59d3\u540d!"),t.BQk())}function W(e,u){1&e&&t.YNc(0,R,2,0,"ng-container",17),2&e&&t.Q6J("ngIf",u.$implicit.hasError("required"))}function j(e,u){1&e&&(t.ynx(0),t._uU(1,"\u8bf7\u8f93\u5165\u90ae\u7bb1!"),t.BQk())}function I(e,u){1&e&&t.YNc(0,j,2,0,"ng-container",17),2&e&&t.Q6J("ngIf",u.$implicit.hasError("required"))}function S(e,u){1&e&&(t.ynx(0),t._uU(1,"\u8bf7\u786e\u5b9a\u89d2\u8272\u6743\u9650!"),t.BQk())}function G(e,u){1&e&&t.YNc(0,S,2,0,"ng-container",17),2&e&&t.Q6J("ngIf",u.$implicit.hasError("required"))}let V=(()=>{class e{constructor(n,o,r){this.fb=n,this.rs=o,this.msg=r,this.listOfOption=[],this.id=""}ngOnInit(){this.id&&this.rs.get(`user/${this.id}`).subscribe(n=>{this.build(n.data)}),this.build(),this.getRoleList()}build(n){this.group=this.fb.group({username:[(n=n||{}).username||"",[y.kI.required]],name:[n.name||"",[]],email:[n.email||"",[]],roles:[n.roles||[],[]],disabled:[n.disabled||!1,[]]})}submit(){return new Promise((n,o)=>{this.group.valid?this.rs.post(this.id?`user/${this.id}`:"user/create",this.group.value).subscribe(c=>{this.msg.success("\u4fdd\u5b58\u6210\u529f"),n(!0)}):Object.values(this.group.controls).forEach(r=>{r.invalid&&(r.markAsDirty(),r.updateValueAndValidity({onlySelf:!0}),o())})})}getRoleList(){this.rs.get("role/list").subscribe(n=>{const o=[];n.data.filter(r=>{o.push({value:r.name,label:r.name}),this.listOfOption=o})})}}return e.\u0275fac=function(n){return new(n||e)(t.Y36(y.qu),t.Y36(d.s),t.Y36(m.dD))},e.\u0275cmp=t.Xpm({type:e,selectors:[["app-users-edit"]],inputs:{id:"id"},decls:34,vars:26,consts:[["nz-form","",3,"formGroup","ngSubmit"],["nzRequired","","nzFor","username",3,"nzSm","nzXs"],[3,"nzSm","nzXs","nzErrorTip"],["nz-input","","formControlName","username"],["nameErrorTpl",""],["nzFor","name",3,"nzSm","nzXs"],["nz-input","","formControlName","name"],["idErrorTpl",""],["nzFor","email",3,"nzSm","nzXs"],["nz-input","","formControlName","email"],["emailErrorTpl",""],[3,"nzSm","nzXs"],["nzMode","multiple","name","roles","formControlName","roles","nzShowSearch","","nzAllowClear","",3,"nzOptions"],["rolesErrorTpl",""],["nzFor","disabled",3,"nzSm","nzXs"],["nzHasFeedback","",3,"nzSm","nzXs"],["nzCheckedChildren","\u662f","nzUnCheckedChildren","\u5426","name","disabled","formControlName","disabled"],[4,"ngIf"]],template:function(n,o){if(1&n&&(t.TgZ(0,"form",0),t.NdJ("ngSubmit",function(){return o.submit()}),t.TgZ(1,"nz-form-item")(2,"nz-form-label",1),t._uU(3,"\u7528\u6237\u540d"),t.qZA(),t.TgZ(4,"nz-form-control",2),t._UZ(5,"input",3),t.YNc(6,Q,1,1,"ng-template",null,4,t.W1O),t.qZA()(),t.TgZ(8,"nz-form-item")(9,"nz-form-label",5),t._uU(10,"\u59d3\u540d"),t.qZA(),t.TgZ(11,"nz-form-control",2),t._UZ(12,"input",6),t.YNc(13,W,1,1,"ng-template",null,7,t.W1O),t.qZA()(),t.TgZ(15,"nz-form-item")(16,"nz-form-label",8),t._uU(17,"\u90ae\u7bb1"),t.qZA(),t.TgZ(18,"nz-form-control",2),t._UZ(19,"input",9),t.YNc(20,I,1,1,"ng-template",null,10,t.W1O),t.qZA()(),t.TgZ(22,"nz-form-item")(23,"nz-form-label",11),t._uU(24,"\u89d2\u8272"),t.qZA(),t.TgZ(25,"nz-form-control",2),t._UZ(26,"nz-select",12),t.YNc(27,G,1,1,"ng-template",null,13,t.W1O),t.qZA()(),t.TgZ(29,"nz-form-item")(30,"nz-form-label",14),t._uU(31,"\u662f\u5426\u7981\u7528"),t.qZA(),t.TgZ(32,"nz-form-control",15),t._UZ(33,"nz-switch",16),t.qZA()()()),2&n){const r=t.MAs(7),c=t.MAs(14),J=t.MAs(21),q=t.MAs(28);t.Q6J("formGroup",o.group),t.xp6(2),t.Q6J("nzSm",6)("nzXs",24),t.xp6(2),t.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",r),t.xp6(5),t.Q6J("nzSm",6)("nzXs",24),t.xp6(2),t.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",c),t.xp6(5),t.Q6J("nzSm",6)("nzXs",24),t.xp6(2),t.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",J),t.xp6(5),t.Q6J("nzSm",6)("nzXs",24),t.xp6(2),t.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",q),t.xp6(1),t.Q6J("nzOptions",o.listOfOption),t.xp6(4),t.Q6J("nzSm",6)("nzXs",24),t.xp6(2),t.Q6J("nzSm",14)("nzXs",24)}},dependencies:[p.O5,z.i,U.Vq,y._Y,y.Fj,y.JJ,y.JL,y.sg,y.u,E.t3,E.SK,D.Lr,D.Nx,D.iK,D.Fd,B.Zp]}),e})();var h=a(3910),l=a(7423),i=a(6672),_=a(6497),s=a(6616),g=a(7044),f=a(1664),x=a(5227),M=a(297),v=a(2451),A=a(6903),Z=a(269),Y=a(2577),K=a(5545);function X(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"button",17),t.NdJ("click",function(){t.CHM(n);const r=t.oxw();return t.KtG(r.reload())}),t._UZ(1,"i",18),t._uU(2," \u5237\u65b0 "),t.qZA()}if(2&e){const n=t.oxw();t.xp6(1),t.Q6J("nzSpin",n.loading)}}function L(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"app-search-box",19),t.NdJ("onSearch",function(r){t.CHM(n);const c=t.oxw();return t.KtG(c.search(r))}),t.qZA()}}function k(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"button",17),t.NdJ("click",function(){t.CHM(n);const r=t.oxw();return t.KtG(r.handleEdit())}),t._UZ(1,"i",20),t._uU(2," \u65b0\u589e "),t.qZA()}}function w(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"button",21),t._UZ(1,"span",22),t.TgZ(2,"a",23),t.NdJ("click",function(){t.CHM(n);const r=t.oxw();return t.KtG(r.handleExport())}),t._uU(3,"\u5bfc\u51fa"),t.qZA()()}if(2&e){const n=t.oxw();t.xp6(2),t.Q6J("href",n.href,t.LSH)}}function nt(e,u){1&e&&t._UZ(0,"app-import",24),2&e&&t.Q6J("url","user/import")}function tt(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"button",25),t.NdJ("click",function(){t.CHM(n);const r=t.oxw();return t.KtG(r.handleBatchDel())}),t._uU(1," \u6279\u91cf\u5220\u9664 "),t.qZA()}}function ot(e,u){1&e&&t._uU(0),2&e&&t.hij("\u603b\u5171 ",u.$implicit," \u6761")}function rt(e,u){1&e&&(t.TgZ(0,"nz-tag",34),t._uU(1," \u7981\u7528 "),t.qZA())}function it(e,u){1&e&&(t.TgZ(0,"nz-tag",35),t._uU(1," \u542f\u7528 "),t.qZA())}function st(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"tr")(1,"td",26),t.NdJ("nzCheckedChange",function(r){const J=t.CHM(n).$implicit,q=t.oxw();return t.KtG(q.handleItemChecked(J.id,r))}),t.qZA(),t.TgZ(2,"td"),t._uU(3),t.qZA(),t.TgZ(4,"td"),t._uU(5),t.qZA(),t.TgZ(6,"td"),t._uU(7),t.qZA(),t.TgZ(8,"td"),t._uU(9),t.qZA(),t.TgZ(10,"td"),t._uU(11),t.qZA(),t.TgZ(12,"td"),t.YNc(13,rt,2,0,"nz-tag",27),t.YNc(14,it,2,0,"nz-tag",28),t.qZA(),t.TgZ(15,"td"),t._uU(16),t.ALo(17,"date"),t.qZA(),t.TgZ(18,"td")(19,"a",29),t.NdJ("click",function(){const c=t.CHM(n).$implicit,J=t.oxw();return t.KtG(J.handleEdit(c.id))}),t._UZ(20,"i",30),t.qZA(),t._UZ(21,"nz-divider",31),t.TgZ(22,"a",32),t.NdJ("nzOnConfirm",function(){const c=t.CHM(n).$implicit,J=t.oxw();return t.KtG(J.delete(c.id))})("nzOnCancel",function(){t.CHM(n);const r=t.oxw();return t.KtG(r.cancel())}),t._UZ(23,"i",33),t.qZA()()()}if(2&e){const n=u.$implicit,o=t.oxw();t.xp6(1),t.Q6J("nzChecked",o.setOfCheckedId.has(n.id)),t.xp6(2),t.Oqu(n.id),t.xp6(2),t.Oqu(n.username),t.xp6(2),t.Oqu(n.name),t.xp6(2),t.Oqu(n.email),t.xp6(2),t.Oqu(n.roles),t.xp6(2),t.Q6J("ngIf",n.disabled),t.xp6(1),t.Q6J("ngIf",!n.disabled),t.xp6(2),t.Oqu(t.lcZ(17,9,n.created))}}let at=(()=>{class e{constructor(n,o,r,c){this.modal=n,this.router=o,this.rs=r,this.msg=c,this.loading=!0,this.uploading=!1,this.datum=[],this.total=1,this.pageSize=20,this.pageIndex=1,this.query={},this.checked=!1,this.indeterminate=!1,this.setOfCheckedId=new Set,this.delResData=[]}reload(){this.datum=[],this.load()}load(){this.loading=!0,this.rs.post("user/search",this.query).subscribe(n=>{this.datum=n.data||[],this.total=n.total,this.setOfCheckedId.clear(),(0,C.oR)(this)}).add(()=>{this.loading=!1})}create(){let n="/user/create";location.pathname.startsWith("/admin")&&(n="/admin"+n),this.router.navigateByUrl(n)}delete(n,o){this.rs.get(`user/${n}/delete`).subscribe(r=>{o?o&&(this.delResData.push(r),o===this.delResData.length&&(this.msg.success("\u5220\u9664\u6210\u529f"),this.load())):(this.msg.success("\u5220\u9664\u6210\u529f"),this.datum=this.datum.filter(c=>c.id!==n))})}handleExport(){this.href="/api/user/export"}onQuery(n){(0,F.r)(n,this.query),this.load()}pageIndexChange(n){this.query.skip=n-1}pageSizeChange(n){this.query.limit=n}search(n){this.query.keyword={name:n},this.query.skip=0,this.load()}handleEdit(n){const r=this.modal.create({nzTitle:n?"\u7f16\u8f91\u7528\u6237":"\u65b0\u589e\u7528\u6237",nzStyle:{top:"20px"},nzContent:V,nzComponentParams:{id:n},nzFooter:[{label:"\u53d6\u6d88",onClick:()=>{r.destroy()}},{label:"\u4fdd\u5b58",type:"primary",onClick:c=>{c.submit().then(()=>{r.destroy(),this.load()},()=>{})}}]})}cancel(){this.msg.info("\u53d6\u6d88\u64cd\u4f5c")}getTableHeight(){return(0,C.NC)(this)}handleBatchDel(){(0,C.mK)(this)}handleAllChecked(n){(0,C.Yk)(n,this)}handleItemChecked(n,o){(0,C.mp)(n,o,this)}}return e.\u0275fac=function(n){return new(n||e)(t.Y36(h.Sf),t.Y36(b.F0),t.Y36(d.s),t.Y36(m.dD))},e.\u0275cmp=t.Xpm({type:e,selectors:[["app-users"]],decls:33,vars:13,consts:[["nz-button","","nzType","primary",3,"click",4,"nzSpaceItem"],[3,"onSearch",4,"nzSpaceItem"],["nz-button","","nzType","primary","class","btn",4,"nzSpaceItem"],[3,"url",4,"nzSpaceItem"],["nz-button","","nzType","primary","nzDanger","",3,"click",4,"nzSpaceItem"],["totalTemplate",""],["nzShowPagination","","nzShowSizeChanger","",3,"nzData","nzLoading","nzFrontPagination","nzTotal","nzShowTotal","nzPageSize","nzPageIndex","nzPageSizeChange","nzPageIndexChange","nzQueryParams"],["basicTable",""],[3,"nzChecked","nzIndeterminate","nzCheckedChange"],["nzColumnKey","id",3,"nzSortFn"],["nzColumnKey","username",3,"nzSortFn"],["nzColumnKey","name"],["nzColumnKey","email"],["nzColumnKey","roles"],["nzColumnKey","disabled"],["nzColumnKey","created",3,"nzSortFn"],[4,"ngFor","ngForOf"],["nz-button","","nzType","primary",3,"click"],["nz-icon","","nzType","reload",3,"nzSpin"],[3,"onSearch"],["nz-icon","","nzType","plus"],["nz-button","","nzType","primary",1,"btn"],["nz-icon","","nzType","cloud-download","nzTheme","outline"],["download","filename",3,"href","click"],[3,"url"],["nz-button","","nzType","primary","nzDanger","",3,"click"],[3,"nzChecked","nzCheckedChange"],["nzColor","error",4,"ngIf"],["nzColor","success",4,"ngIf"],[3,"click"],["nz-icon","","nzType","edit"],["nzType","vertical"],["nz-popconfirm","","nzPopconfirmTitle","\u786e\u5b9a\u5220\u9664?","nzPopconfirmPlacement","topLeft",3,"nzOnConfirm","nzOnCancel"],["nz-icon","","nzType","delete"],["nzColor","error"],["nzColor","success"]],template:function(n,o){if(1&n&&(t.TgZ(0,"app-toolbar")(1,"nz-space"),t.YNc(2,X,3,1,"button",0),t.YNc(3,L,1,0,"app-search-box",1),t.YNc(4,k,3,0,"button",0),t.YNc(5,w,4,1,"button",2),t.YNc(6,nt,1,1,"app-import",3),t.YNc(7,tt,2,0,"button",4),t.qZA()(),t.YNc(8,ot,1,1,"ng-template",null,5,t.W1O),t.TgZ(10,"nz-table",6,7),t.NdJ("nzPageSizeChange",function(c){return o.pageSizeChange(c)})("nzPageIndexChange",function(c){return o.pageIndexChange(c)})("nzQueryParams",function(c){return o.onQuery(c)}),t.TgZ(12,"thead")(13,"tr")(14,"th",8),t.NdJ("nzCheckedChange",function(c){return o.handleAllChecked(c)}),t.qZA(),t.TgZ(15,"th",9),t._uU(16,"ID"),t.qZA(),t.TgZ(17,"th",10),t._uU(18,"\u7528\u6237\u540d"),t.qZA(),t.TgZ(19,"th",11),t._uU(20,"\u540d\u79f0"),t.qZA(),t.TgZ(21,"th",12),t._uU(22,"\u90ae\u7bb1"),t.qZA(),t.TgZ(23,"th",13),t._uU(24,"\u89d2\u8272"),t.qZA(),t.TgZ(25,"th",14),t._uU(26,"\u72b6\u6001"),t.qZA(),t.TgZ(27,"th",15),t._uU(28,"\u65e5\u671f"),t.qZA(),t.TgZ(29,"th"),t._uU(30,"\u64cd\u4f5c"),t.qZA()()(),t.TgZ(31,"tbody"),t.YNc(32,st,24,11,"tr",16),t.qZA()()),2&n){const r=t.MAs(9),c=t.MAs(11);t.xp6(10),t.Q6J("nzData",o.datum)("nzLoading",o.loading)("nzFrontPagination",!1)("nzTotal",o.total)("nzShowTotal",r)("nzPageSize",o.pageSize)("nzPageIndex",o.pageIndex),t.xp6(4),t.Q6J("nzChecked",o.checked)("nzIndeterminate",o.indeterminate),t.xp6(1),t.Q6J("nzSortFn",!0),t.xp6(2),t.Q6J("nzSortFn",!0),t.xp6(10),t.Q6J("nzSortFn",!0),t.xp6(5),t.Q6J("ngForOf",c.data)}},dependencies:[p.sg,p.O5,l.Ls,i.j,_.JW,s.ix,g.w,f.dQ,x.n,M.B,v.k,A.NU,A.$1,Z.N8,Z.qD,Z.Uo,Z._C,Z.h7,Z.Om,Z.p0,Z.$Z,Z.g6,Y.g,K.u],styles:[".btn[_ngcontent-%COMP%]{display:flex;justify-content:center;align-items:center;width:82px;height:32px;border-radius:3px}.btn[_ngcontent-%COMP%]   a[_ngcontent-%COMP%]{color:#fff;margin-left:5px;font-size:14px}"]}),e})(),lt=(()=>{class e{}return e.\u0275fac=function(n){return new(n||e)},e.\u0275cmp=t.Xpm({type:e,selectors:[["app-user-detail"]],decls:2,vars:0,template:function(n,o){1&n&&(t.TgZ(0,"p"),t._uU(1,"user-detail works!"),t.qZA())}}),e})();function ut(e,u){1&e&&(t.ynx(0),t._uU(1,"\u8bf7\u8f93\u5165id!"),t.BQk())}function ct(e,u){1&e&&t.YNc(0,ut,2,0,"ng-container",1),2&e&&t.Q6J("ngIf",u.$implicit.hasError("required"))}function dt(e,u){if(1&e&&(t.TgZ(0,"nz-form-item")(1,"nz-form-label",9),t._uU(2,"ID"),t.qZA(),t.TgZ(3,"nz-form-control",3),t._UZ(4,"input",10),t.YNc(5,ct,1,1,"ng-template",null,11,t.W1O),t.qZA()()),2&e){const n=t.MAs(6);t.xp6(1),t.Q6J("nzSm",6)("nzXs",24),t.xp6(2),t.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",n)}}function pt(e,u){1&e&&(t.ynx(0),t._uU(1,"\u8bf7\u8f93\u5165\u540d\u79f0!"),t.BQk())}function mt(e,u){1&e&&t.YNc(0,pt,2,0,"ng-container",1),2&e&&t.Q6J("ngIf",u.$implicit.hasError("required"))}function ht(e,u){1&e&&(t.ynx(0),t._uU(1,"\u8bf7\u786e\u5b9a\u89d2\u8272\u6743\u9650!"),t.BQk())}function _t(e,u){1&e&&t.YNc(0,ht,2,0,"ng-container",1),2&e&&t.Q6J("ngIf",u.$implicit.hasError("required"))}let et=(()=>{class e{constructor(n,o,r,c,J,q){this.fb=n,this.router=o,this.route=r,this.rs=c,this.msg=J,this.datePipe=q,this.id="",this.listOfOption=[],this.listOfSelectedValue=[]}ngOnInit(){this.build(),this.id&&this.rs.get(`role/${this.id}`).subscribe(n=>{this.build(n.data)}),this.getRoleList()}build(n){n=n||{};const{name:o,id:r,privileges:c}=n||{};this.group=this.fb.group({name:[o||"",[y.kI.required]],id:[r||"",[y.kI.required]],privileges:[c||[],[y.kI.required]]})}getRoleList(){this.rs.get("privileges").subscribe(n=>{const{data:o}=n,r=[];for(const c in o)r.push({label:o[c],value:c});this.listOfOption=r})}submit(){return new Promise((n,o)=>{this.group.valid?this.rs.post(this.id?`role/${this.id}`:"role/create",this.group.value).subscribe(c=>{this.msg.success("\u4fdd\u5b58\u6210\u529f"),n(!0)}):Object.values(this.group.controls).forEach(r=>{r.invalid&&(r.markAsDirty(),r.updateValueAndValidity({onlySelf:!0}),o())})})}handleCancel(){const n=`${(0,C.kh)()}/user/role`;this.router.navigateByUrl(n)}}return e.\u0275fac=function(n){return new(n||e)(t.Y36(y.qu),t.Y36(b.F0),t.Y36(b.gz),t.Y36(d.s),t.Y36(m.dD),t.Y36(p.uU))},e.\u0275cmp=t.Xpm({type:e,selectors:[["app-role-edit"]],inputs:{id:"id"},features:[t._Bn([p.uU])],decls:16,vars:13,consts:[["nz-form","",3,"formGroup","ngSubmit"],[4,"ngIf"],["nzRequired","","nzFor","name",3,"nzSm","nzXs"],[3,"nzSm","nzXs","nzErrorTip"],["nz-input","","formControlName","name"],["nameErrorTpl",""],["nzRequired","",3,"nzSm","nzXs"],["nzMode","multiple","name","privileges","formControlName","privileges","nzShowSearch","","nzAllowClear","",3,"nzOptions"],["privilegesErrorTpl",""],["nzFor","id","nzRequired","",3,"nzSm","nzXs"],["nz-input","","formControlName","id"],["idErrorTpl",""]],template:function(n,o){if(1&n&&(t.TgZ(0,"form",0),t.NdJ("ngSubmit",function(){return o.submit()}),t.YNc(1,dt,7,5,"nz-form-item",1),t.TgZ(2,"nz-form-item")(3,"nz-form-label",2),t._uU(4,"\u540d\u79f0"),t.qZA(),t.TgZ(5,"nz-form-control",3),t._UZ(6,"input",4),t.YNc(7,mt,1,1,"ng-template",null,5,t.W1O),t.qZA()(),t.TgZ(9,"nz-form-item")(10,"nz-form-label",6),t._uU(11,"\u6743\u9650"),t.qZA(),t.TgZ(12,"nz-form-control",3),t._UZ(13,"nz-select",7),t.YNc(14,_t,1,1,"ng-template",null,8,t.W1O),t.qZA()()()),2&n){const r=t.MAs(8),c=t.MAs(15);t.Q6J("formGroup",o.group),t.xp6(1),t.Q6J("ngIf",!o.id),t.xp6(2),t.Q6J("nzSm",6)("nzXs",24),t.xp6(2),t.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",r),t.xp6(5),t.Q6J("nzSm",6)("nzXs",24),t.xp6(2),t.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",c),t.xp6(1),t.Q6J("nzOptions",o.listOfOption)}},dependencies:[p.O5,U.Vq,y._Y,y.Fj,y.JJ,y.JL,y.sg,y.u,E.t3,E.SK,D.Lr,D.Nx,D.iK,D.Fd,B.Zp]}),e})(),ft=(()=>{class e{constructor(n){this.rs=n,this.range=4e4,(!localStorage.getItem("roleObj")||!this.judgelocalStorage("roleObj"))&&this.rs.get("privileges").subscribe(r=>{this.setlocalStorage("roleObj",r.data||[])})}transform(n,...o){const r=[];if(localStorage.getItem("roleObj")){const c=localStorage.getItem("roleObj")||"",J=JSON.parse(c);n.forEach(q=>{J[q]&&r.push(J[q])})}return r.join(",")}setlocalStorage(n,o){const r=(new Date).getTime();o.date=r,localStorage.setItem(n,JSON.stringify(o))}judgelocalStorage(n){const o=(new Date).getTime(),r=JSON.parse(localStorage.getItem(n)||"");return o-r.time>this.range?(r.removeItem(n),null):r.key}}return e.\u0275fac=function(n){return new(n||e)(t.Y36(d.s,16))},e.\u0275pipe=t.Yjl({name:"handlePrivileges",type:e,pure:!0}),e})();function gt(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"button",14),t.NdJ("click",function(){t.CHM(n);const r=t.oxw();return t.KtG(r.reload())}),t._UZ(1,"i",15),t._uU(2," \u5237\u65b0 "),t.qZA()}if(2&e){const n=t.oxw();t.xp6(1),t.Q6J("nzSpin",n.loading)}}function zt(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"app-search-box",16),t.NdJ("onSearch",function(r){t.CHM(n);const c=t.oxw();return t.KtG(c.search(r))}),t.qZA()}}function Ct(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"button",14),t.NdJ("click",function(){t.CHM(n);const r=t.oxw();return t.KtG(r.handleEdit())}),t._UZ(1,"i",17),t._uU(2," \u65b0\u589e "),t.qZA()}}function Tt(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"button",18),t._UZ(1,"span",19),t.TgZ(2,"a",20),t.NdJ("click",function(){t.CHM(n);const r=t.oxw();return t.KtG(r.handleExport())}),t._uU(3,"\u5bfc\u51fa"),t.qZA()()}if(2&e){const n=t.oxw();t.xp6(2),t.Q6J("href",n.href,t.LSH)}}function xt(e,u){1&e&&t._UZ(0,"app-import",21),2&e&&t.Q6J("url","role/import")}function yt(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"button",22),t.NdJ("click",function(){t.CHM(n);const r=t.oxw();return t.KtG(r.handleBatchDel())}),t._uU(1," \u6279\u91cf\u5220\u9664 "),t.qZA()}}function St(e,u){1&e&&t._uU(0),2&e&&t.hij("\u603b\u5171 ",u.$implicit," \u6761")}function Zt(e,u){if(1&e){const n=t.EpF();t.TgZ(0,"tr")(1,"td",23),t.NdJ("nzCheckedChange",function(r){const J=t.CHM(n).$implicit,q=t.oxw();return t.KtG(q.handleItemChecked(J.id,r))}),t.qZA(),t.TgZ(2,"td"),t._uU(3),t.qZA(),t.TgZ(4,"td"),t._uU(5),t.qZA(),t.TgZ(6,"td"),t._uU(7),t.ALo(8,"handlePrivileges"),t.qZA(),t.TgZ(9,"td"),t._uU(10),t.ALo(11,"date"),t.qZA(),t.TgZ(12,"td")(13,"a",24),t.NdJ("click",function(){const c=t.CHM(n).$implicit,J=t.oxw();return t.KtG(J.handleEdit(c.id))}),t._UZ(14,"i",25),t.qZA(),t._UZ(15,"nz-divider",26),t.TgZ(16,"a",27),t.NdJ("nzOnConfirm",function(){const c=t.CHM(n).$implicit,J=t.oxw();return t.KtG(J.delete(c.id))})("nzOnCancel",function(){t.CHM(n);const r=t.oxw();return t.KtG(r.cancel())}),t._UZ(17,"i",28),t.qZA()()()}if(2&e){const n=u.$implicit,o=t.oxw();t.xp6(1),t.Q6J("nzChecked",o.setOfCheckedId.has(n.id)),t.xp6(2),t.Oqu(n.id),t.xp6(2),t.Oqu(n.name),t.xp6(2),t.Oqu(t.lcZ(8,5,n.privileges)),t.xp6(3),t.Oqu(t.lcZ(11,7,n.created))}}const Ut=[{path:"",pathMatch:"full",redirectTo:"list"},{path:"list",component:at},{path:"detail/:id",component:lt},{path:"edit/:id",component:V},{path:"create",component:V},{path:"role/create",component:et},{path:"role",component:(()=>{class e{constructor(n,o,r,c){this.router=n,this.rs=o,this.modal=r,this.msg=c,this.loading=!0,this.uploading=!1,this.datum=[],this.total=1,this.pageSize=20,this.pageIndex=1,this.query={},this.isVisible=!1,this.roleObj={},this.checked=!1,this.indeterminate=!1,this.setOfCheckedId=new Set,this.delResData=[],this.load()}reload(){this.datum=[],this.load()}load(){this.loading=!0,this.rs.post("role/search",this.query).subscribe(n=>{this.datum=n.data||[],this.total=n.total,this.setOfCheckedId.clear(),(0,C.oR)(this)}).add(()=>{this.loading=!1})}delete(n,o){this.rs.get(`role/${n}/delete`).subscribe(r=>{o?o&&(this.delResData.push(r),o===this.delResData.length&&(this.msg.success("\u5220\u9664\u6210\u529f"),this.load())):(this.msg.success("\u5220\u9664\u6210\u529f"),this.datum=this.datum.filter(c=>c.id!==n))})}handleExport(){this.href="/api/role/export"}onQuery(n){(0,F.r)(n,this.query),this.load()}pageIndexChange(n){this.query.skip=n-1}pageSizeChange(n){this.query.limit=n}search(n){this.query.keyword={name:n},this.query.skip=0,this.load()}handleEdit(n){const r=this.modal.create({nzTitle:n?"\u7f16\u8f91\u89d2\u8272":"\u65b0\u589e\u89d2\u8272",nzStyle:{top:"20px"},nzContent:et,nzComponentParams:{id:n},nzFooter:[{label:"\u53d6\u6d88",onClick:()=>{r.destroy()}},{label:"\u4fdd\u5b58",type:"primary",onClick:c=>{c.submit().then(()=>{r.destroy(),this.load()},()=>{})}}]})}cancel(){this.msg.info("\u70b9\u51fb\u53d6\u6d88")}getTableHeight(){return(0,C.NC)(this)}handleBatchDel(){(0,C.mK)(this)}handleAllChecked(n){(0,C.Yk)(n,this)}handleItemChecked(n,o){(0,C.mp)(n,o,this)}}return e.\u0275fac=function(n){return new(n||e)(t.Y36(b.F0),t.Y36(d.s),t.Y36(h.Sf),t.Y36(m.dD))},e.\u0275cmp=t.Xpm({type:e,selectors:[["app-role"]],decls:27,vars:12,consts:[["nz-button","","nzType","primary",3,"click",4,"nzSpaceItem"],[3,"onSearch",4,"nzSpaceItem"],["nz-button","","nzType","primary","class","btn",4,"nzSpaceItem"],[3,"url",4,"nzSpaceItem"],["nz-button","","nzType","primary","nzDanger","",3,"click",4,"nzSpaceItem"],["totalTemplate",""],["nzShowPagination","","nzShowSizeChanger","",3,"nzData","nzLoading","nzFrontPagination","nzTotal","nzShowTotal","nzPageSize","nzPageIndex","nzPageSizeChange","nzPageIndexChange","nzQueryParams"],["basicTable",""],[3,"nzChecked","nzIndeterminate","nzCheckedChange"],["nzColumnKey","id",3,"nzSortFn"],["nzColumnKey","name"],["nzColumnKey","privileges"],["nzColumnKey","created",3,"nzSortFn"],[4,"ngFor","ngForOf"],["nz-button","","nzType","primary",3,"click"],["nz-icon","","nzType","reload",3,"nzSpin"],[3,"onSearch"],["nz-icon","","nzType","plus"],["nz-button","","nzType","primary",1,"btn"],["nz-icon","","nzType","cloud-download","nzTheme","outline"],["download","filename",3,"href","click"],[3,"url"],["nz-button","","nzType","primary","nzDanger","",3,"click"],[3,"nzChecked","nzCheckedChange"],[3,"click"],["nz-icon","","nzType","edit"],["nzType","vertical"],["nz-popconfirm","","nzPopconfirmTitle","\u786e\u5b9a\u5220\u9664?","nzPopconfirmPlacement","topLeft",3,"nzOnConfirm","nzOnCancel"],["nz-icon","","nzType","delete"]],template:function(n,o){if(1&n&&(t.TgZ(0,"app-toolbar")(1,"nz-space"),t.YNc(2,gt,3,1,"button",0),t.YNc(3,zt,1,0,"app-search-box",1),t.YNc(4,Ct,3,0,"button",0),t.YNc(5,Tt,4,1,"button",2),t.YNc(6,xt,1,1,"app-import",3),t.YNc(7,yt,2,0,"button",4),t.qZA()(),t.YNc(8,St,1,1,"ng-template",null,5,t.W1O),t.TgZ(10,"nz-table",6,7),t.NdJ("nzPageSizeChange",function(c){return o.pageSizeChange(c)})("nzPageIndexChange",function(c){return o.pageIndexChange(c)})("nzQueryParams",function(c){return o.onQuery(c)}),t.TgZ(12,"thead")(13,"tr")(14,"th",8),t.NdJ("nzCheckedChange",function(c){return o.handleAllChecked(c)}),t.qZA(),t.TgZ(15,"th",9),t._uU(16,"id"),t.qZA(),t.TgZ(17,"th",10),t._uU(18,"\u540d\u79f0"),t.qZA(),t.TgZ(19,"th",11),t._uU(20,"\u6743\u9650"),t.qZA(),t.TgZ(21,"th",12),t._uU(22,"\u521b\u5efa\u65f6\u95f4"),t.qZA(),t.TgZ(23,"th"),t._uU(24,"\u64cd\u4f5c"),t.qZA()()(),t.TgZ(25,"tbody"),t.YNc(26,Zt,18,9,"tr",13),t.qZA()()),2&n){const r=t.MAs(9),c=t.MAs(11);t.xp6(10),t.Q6J("nzData",o.datum)("nzLoading",o.loading)("nzFrontPagination",!1)("nzTotal",o.total)("nzShowTotal",r)("nzPageSize",o.pageSize)("nzPageIndex",o.pageIndex),t.xp6(4),t.Q6J("nzChecked",o.checked)("nzIndeterminate",o.indeterminate),t.xp6(1),t.Q6J("nzSortFn",!0),t.xp6(6),t.Q6J("nzSortFn",!0),t.xp6(5),t.Q6J("ngForOf",c.data)}},dependencies:[p.sg,l.Ls,_.JW,s.ix,g.w,f.dQ,x.n,M.B,v.k,A.NU,A.$1,Z.N8,Z.qD,Z.Uo,Z._C,Z.h7,Z.Om,Z.p0,Z.$Z,Z.g6,Y.g,K.u,ft],styles:[".btn[_ngcontent-%COMP%]{display:flex;justify-content:center;align-items:center;width:82px;height:32px;border-radius:3px}.btn[_ngcontent-%COMP%]   a[_ngcontent-%COMP%]{color:#fff;margin-left:5px;font-size:14px}"]}),e})()},{path:"privillege/:id",component:et},{path:"me",component:T},{path:"password",component:$.I},{path:"**",component:O.r}];let Ot=(()=>{class e{}return e.\u0275fac=function(n){return new(n||e)},e.\u0275mod=t.oAB({type:e}),e.\u0275inj=t.cJS({imports:[b.Bz.forChild(Ut),b.Bz]}),e})();var vt=a(8284),bt=a(3325),Mt=a(1971),At=a(7096),Et=a(9166);let Dt=(()=>{class e{}return e.\u0275fac=function(n){return new(n||e)},e.\u0275mod=t.oAB({type:e}),e.\u0275inj=t.cJS({imports:[p.ez,Ot,vt.wm,bt.ip,l.PV,i.X,_._p,z.m,s.sL,Mt.vh,h.Qp,U.LV,y.u5,y.UX,D.U5,B.o7,At.Zf,Et.Y,A.zf,Z.HQ,Y.S]}),e})()},235:(H,P,a)=>{"use strict";function p(){return location.pathname.startsWith("/admin")?"/admin":""}function t(d){const m=document.querySelector(".ant-table")?.getBoundingClientRect().top||0;return{y:(document.querySelector(".ant-layout")?.clientHeight||0)-m-120+"px"}}function T(d,m){m.datum&&(m.datum.forEach(z=>F(z.id,d,m)),C(m))}function $(d,m,z){F(d,m,z),C(z)}function F(d,m,z){m?z.setOfCheckedId.add(d):z.setOfCheckedId.delete(d)}function C(d){d.datum&&(d.checked=d.datum.every(m=>d.setOfCheckedId.has(m.id)),d.indeterminate=d.datum.some(m=>d.setOfCheckedId.has(m.id))&&!d.checked)}function y(d){d.delResData=[];const m=d.setOfCheckedId.size;if(!m)return void d.msg.warning("\u8bf7\u5148\u52fe\u9009\u5220\u9664\u9879");const z=Array.from(d.setOfCheckedId);d.modal.confirm({nzTitle:`\u786e\u5b9a\u5220\u9664\u52fe\u9009\u7684${m}\u9879\uff1f`,nzOnOk:()=>{z.forEach(U=>{d.delete(U,m)})}})}a.d(P,{NC:()=>t,Yk:()=>T,kh:()=>p,mK:()=>y,mp:()=>$,oR:()=>C})},1764:function(H){H.exports=function(){"use strict";var a=6e4,p=36e5,b="millisecond",O="second",t="minute",T="hour",$="day",F="week",C="month",y="quarter",d="year",m="date",z="Invalid Date",U=/^(\d{4})[-/]?(\d{1,2})?[-/]?(\d{0,2})[Tt\s]*(\d{1,2})?:?(\d{1,2})?:?(\d{1,2})?[.:]?(\d+)?$/,E=/\[([^\]]+)]|Y{1,4}|M{1,4}|D{1,2}|d{1,4}|H{1,2}|h{1,2}|a|A|m{1,2}|s{1,2}|Z{1,2}|SSS/g,D={name:"en",weekdays:"Sunday_Monday_Tuesday_Wednesday_Thursday_Friday_Saturday".split("_"),months:"January_February_March_April_May_June_July_August_September_October_November_December".split("_"),ordinal:function(h){var l=["th","st","nd","rd"],i=h%100;return"["+h+(l[(i-20)%10]||l[i]||l[0])+"]"}},B=function(h,l,i){var _=String(h);return!_||_.length>=l?h:""+Array(l+1-_.length).join(i)+h},N={s:B,z:function(h){var l=-h.utcOffset(),i=Math.abs(l),_=Math.floor(i/60),s=i%60;return(l<=0?"+":"-")+B(_,2,"0")+":"+B(s,2,"0")},m:function h(l,i){if(l.date()<i.date())return-h(i,l);var _=12*(i.year()-l.year())+(i.month()-l.month()),s=l.clone().add(_,C),g=i-s<0,f=l.clone().add(_+(g?-1:1),C);return+(-(_+(i-s)/(g?s-f:f-s))||0)},a:function(h){return h<0?Math.ceil(h)||0:Math.floor(h)},p:function(h){return{M:C,y:d,w:F,d:$,D:m,h:T,m:t,s:O,ms:b,Q:y}[h]||String(h||"").toLowerCase().replace(/s$/,"")},u:function(h){return void 0===h}},Q="en",R={};R[Q]=D;var W=function(h){return h instanceof G},j=function h(l,i,_){var s;if(!l)return Q;if("string"==typeof l){var g=l.toLowerCase();R[g]&&(s=g),i&&(R[g]=i,s=g);var f=l.split("-");if(!s&&f.length>1)return h(f[0])}else{var x=l.name;R[x]=l,s=x}return!_&&s&&(Q=s),s||!_&&Q},I=function(h,l){if(W(h))return h.clone();var i="object"==typeof l?l:{};return i.date=h,i.args=arguments,new G(i)},S=N;S.l=j,S.i=W,S.w=function(h,l){return I(h,{locale:l.$L,utc:l.$u,x:l.$x,$offset:l.$offset})};var G=function(){function h(i){this.$L=j(i.locale,null,!0),this.parse(i)}var l=h.prototype;return l.parse=function(i){this.$d=function(_){var s=_.date,g=_.utc;if(null===s)return new Date(NaN);if(S.u(s))return new Date;if(s instanceof Date)return new Date(s);if("string"==typeof s&&!/Z$/i.test(s)){var f=s.match(U);if(f){var x=f[2]-1||0,M=(f[7]||"0").substring(0,3);return g?new Date(Date.UTC(f[1],x,f[3]||1,f[4]||0,f[5]||0,f[6]||0,M)):new Date(f[1],x,f[3]||1,f[4]||0,f[5]||0,f[6]||0,M)}}return new Date(s)}(i),this.$x=i.x||{},this.init()},l.init=function(){var i=this.$d;this.$y=i.getFullYear(),this.$M=i.getMonth(),this.$D=i.getDate(),this.$W=i.getDay(),this.$H=i.getHours(),this.$m=i.getMinutes(),this.$s=i.getSeconds(),this.$ms=i.getMilliseconds()},l.$utils=function(){return S},l.isValid=function(){return this.$d.toString()!==z},l.isSame=function(i,_){var s=I(i);return this.startOf(_)<=s&&s<=this.endOf(_)},l.isAfter=function(i,_){return I(i)<this.startOf(_)},l.isBefore=function(i,_){return this.endOf(_)<I(i)},l.$g=function(i,_,s){return S.u(i)?this[_]:this.set(s,i)},l.unix=function(){return Math.floor(this.valueOf()/1e3)},l.valueOf=function(){return this.$d.getTime()},l.startOf=function(i,_){var s=this,g=!!S.u(_)||_,f=S.p(i),x=function(L,k){var w=S.w(s.$u?Date.UTC(s.$y,k,L):new Date(s.$y,k,L),s);return g?w:w.endOf($)},M=function(L,k){return S.w(s.toDate()[L].apply(s.toDate("s"),(g?[0,0,0,0]:[23,59,59,999]).slice(k)),s)},v=this.$W,A=this.$M,Z=this.$D,Y="set"+(this.$u?"UTC":"");switch(f){case d:return g?x(1,0):x(31,11);case C:return g?x(1,A):x(0,A+1);case F:var K=this.$locale().weekStart||0,X=(v<K?v+7:v)-K;return x(g?Z-X:Z+(6-X),A);case $:case m:return M(Y+"Hours",0);case T:return M(Y+"Minutes",1);case t:return M(Y+"Seconds",2);case O:return M(Y+"Milliseconds",3);default:return this.clone()}},l.endOf=function(i){return this.startOf(i,!1)},l.$set=function(i,_){var s,g=S.p(i),f="set"+(this.$u?"UTC":""),x=(s={},s[$]=f+"Date",s[m]=f+"Date",s[C]=f+"Month",s[d]=f+"FullYear",s[T]=f+"Hours",s[t]=f+"Minutes",s[O]=f+"Seconds",s[b]=f+"Milliseconds",s)[g],M=g===$?this.$D+(_-this.$W):_;if(g===C||g===d){var v=this.clone().set(m,1);v.$d[x](M),v.init(),this.$d=v.set(m,Math.min(this.$D,v.daysInMonth())).$d}else x&&this.$d[x](M);return this.init(),this},l.set=function(i,_){return this.clone().$set(i,_)},l.get=function(i){return this[S.p(i)]()},l.add=function(i,_){var s,g=this;i=Number(i);var f=S.p(_),x=function(A){var Z=I(g);return S.w(Z.date(Z.date()+Math.round(A*i)),g)};if(f===C)return this.set(C,this.$M+i);if(f===d)return this.set(d,this.$y+i);if(f===$)return x(1);if(f===F)return x(7);var M=(s={},s[t]=a,s[T]=p,s[O]=1e3,s)[f]||1,v=this.$d.getTime()+i*M;return S.w(v,this)},l.subtract=function(i,_){return this.add(-1*i,_)},l.format=function(i){var _=this,s=this.$locale();if(!this.isValid())return s.invalidDate||z;var g=i||"YYYY-MM-DDTHH:mm:ssZ",f=S.z(this),x=this.$H,M=this.$m,v=this.$M,A=s.weekdays,Z=s.months,Y=function(k,w,nt,tt){return k&&(k[w]||k(_,g))||nt[w].slice(0,tt)},K=function(k){return S.s(x%12||12,k,"0")},X=s.meridiem||function(k,w,nt){var tt=k<12?"AM":"PM";return nt?tt.toLowerCase():tt},L={YY:String(this.$y).slice(-2),YYYY:this.$y,M:v+1,MM:S.s(v+1,2,"0"),MMM:Y(s.monthsShort,v,Z,3),MMMM:Y(Z,v),D:this.$D,DD:S.s(this.$D,2,"0"),d:String(this.$W),dd:Y(s.weekdaysMin,this.$W,A,2),ddd:Y(s.weekdaysShort,this.$W,A,3),dddd:A[this.$W],H:String(x),HH:S.s(x,2,"0"),h:K(1),hh:K(2),a:X(x,M,!0),A:X(x,M,!1),m:String(M),mm:S.s(M,2,"0"),s:String(this.$s),ss:S.s(this.$s,2,"0"),SSS:S.s(this.$ms,3,"0"),Z:f};return g.replace(E,function(k,w){return w||L[k]||f.replace(":","")})},l.utcOffset=function(){return 15*-Math.round(this.$d.getTimezoneOffset()/15)},l.diff=function(i,_,s){var g,f=S.p(_),x=I(i),M=(x.utcOffset()-this.utcOffset())*a,v=this-x,A=S.m(this,x);return A=(g={},g[d]=A/12,g[C]=A,g[y]=A/3,g[F]=(v-M)/6048e5,g[$]=(v-M)/864e5,g[T]=v/p,g[t]=v/a,g[O]=v/1e3,g)[f]||v,s?A:S.a(A)},l.daysInMonth=function(){return this.endOf(C).$D},l.$locale=function(){return R[this.$L]},l.locale=function(i,_){if(!i)return this.$L;var s=this.clone(),g=j(i,_,!0);return g&&(s.$L=g),s},l.clone=function(){return S.w(this.$d,this)},l.toDate=function(){return new Date(this.valueOf())},l.toJSON=function(){return this.isValid()?this.toISOString():null},l.toISOString=function(){return this.$d.toISOString()},l.toString=function(){return this.$d.toUTCString()},h}(),V=G.prototype;return I.prototype=V,[["$ms",b],["$s",O],["$m",t],["$H",T],["$W",$],["$M",C],["$y",d],["$D",m]].forEach(function(h){V[h[1]]=function(l){return this.$g(l,h[0],h[1])}}),I.extend=function(h,l){return h.$i||(h(l,G,I),h.$i=!0),I},I.locale=j,I.isDayjs=W,I.unix=function(h){return I(1e3*h)},I.en=R[Q],I.Ls=R,I.p={},I}()}}]);