"use strict";(self.webpackChunkiot_master_ui=self.webpackChunkiot_master_ui||[]).push([[518],{7518:(an,z,i)=>{i.r(z),i.d(z,{PluginModule:()=>en});var d=i(6895),p=i(4575),S=i(5380),l=i(433),n=i(4650),f=i(1445),C=i(9651),u=i(3679),m=i(6704),T=i(5635),v=i(1243),Z=i(8213);function x(e,r){1&e&&(n.ynx(0),n._uU(1,"\u8bf7\u8f93\u5165ID!"),n.BQk())}function b(e,r){1&e&&n.YNc(0,x,2,0,"ng-container",23),2&e&&n.Q6J("ngIf",r.$implicit.hasError("required"))}function E(e,r){1&e&&(n.ynx(0),n._uU(1,"\u8bf7\u8f93\u5165\u540d\u79f0!"),n.BQk())}function F(e,r){1&e&&n.YNc(0,E,2,0,"ng-container",23),2&e&&n.Q6J("ngIf",r.$implicit.hasError("required"))}function J(e,r){1&e&&(n.ynx(0),n._uU(1,"\u8bf7\u8f93\u5165\u7248\u672c!"),n.BQk())}function y(e,r){1&e&&n.YNc(0,J,2,0,"ng-container",23),2&e&&n.Q6J("ngIf",r.$implicit.hasError("required"))}function A(e,r){1&e&&(n.ynx(0),n._uU(1,"\u8bf7\u8f93\u5165\u547d\u4ee4!"),n.BQk())}function Q(e,r){1&e&&n.YNc(0,A,2,0,"ng-container",23),2&e&&n.Q6J("ngIf",r.$implicit.hasError("required"))}let g=(()=>{class e{constructor(t,o,a){this.fb=t,this.rs=o,this.msg=a,this.id=""}ngOnInit(){this.id&&this.rs.get(`plugin/${this.id}`).subscribe(t=>{this.build(t.data)}),this.build()}build(t){this.group=this.fb.group({id:[(t=t||{}).id||"",[]],name:[t.name||"",[l.kI.required]],version:[t.version||"",[]],username:[t.username||"",[]],password:[t.password||"",[]],external:[t.external||!1,[]],command:[t.command||"",[]],dependencies:[t.dependencies||{},[]],disabled:[t.disabled||!1,[]]})}submit(){return new Promise((t,o)=>{this.group.valid?this.rs.post(this.id?`plugin/${this.id}`:"plugin/create",this.group.value).subscribe(s=>{this.msg.success("\u4fdd\u5b58\u6210\u529f"),t(!0)}):Object.values(this.group.controls).forEach(a=>{a.invalid&&(a.markAsDirty(),a.updateValueAndValidity({onlySelf:!0}),o())})})}}return e.\u0275fac=function(t){return new(t||e)(n.Y36(l.qu),n.Y36(f.s),n.Y36(C.dD))},e.\u0275cmp=n.Xpm({type:e,selectors:[["app-plugin-edit"]],inputs:{id:"id"},decls:50,vars:37,consts:[["nz-form","",3,"formGroup","ngSubmit"],["nzFor","id",3,"nzSm","nzXs"],[3,"nzSm","nzXs","nzErrorTip"],["nz-input","","formControlName","id"],["IDErrorTpl",""],["nzFor","name","nzRequired","",3,"nzSm","nzXs"],["nz-input","","formControlName","name"],["nameErrorTpl",""],["nzFor","username",3,"nzSm","nzXs"],[3,"nzSm","nzXs"],["nz-input","","formControlName","username"],["nzFor","password",3,"nzSm","nzXs"],["nz-input","","type","password","formControlName","password"],["nzFor","external",3,"nzSm","nzXs"],["nz-checkbox","","formControlName","external"],["nzFor","version",3,"nzSm","nzXs"],["nz-input","","formControlName","version"],["verErrorTpl",""],["nzFor","command",3,"nzSm","nzXs"],["nz-input","","formControlName","command"],["comErrorTpl",""],["nzFor","disabled",3,"nzSm","nzXs"],["nzCheckedChildren","\u662f","nzUnCheckedChildren","\u5426","name","disabled","formControlName","disabled"],[4,"ngIf"]],template:function(t,o){if(1&t&&(n.TgZ(0,"form",0),n.NdJ("ngSubmit",function(){return o.submit()}),n.TgZ(1,"nz-form-item")(2,"nz-form-label",1),n._uU(3,"ID"),n.qZA(),n.TgZ(4,"nz-form-control",2),n._UZ(5,"input",3),n.YNc(6,b,1,1,"ng-template",null,4,n.W1O),n.qZA()(),n.TgZ(8,"nz-form-item")(9,"nz-form-label",5),n._uU(10,"\u540d\u79f0"),n.qZA(),n.TgZ(11,"nz-form-control",2),n._UZ(12,"input",6),n.YNc(13,F,1,1,"ng-template",null,7,n.W1O),n.qZA()(),n.TgZ(15,"nz-form-item")(16,"nz-form-label",8),n._uU(17,"\u7528\u6237\u540d"),n.qZA(),n.TgZ(18,"nz-form-control",9),n._UZ(19,"input",10),n.qZA()(),n.TgZ(20,"nz-form-item")(21,"nz-form-label",11),n._uU(22,"\u5bc6\u7801"),n.qZA(),n.TgZ(23,"nz-form-control",9),n._UZ(24,"input",12),n.qZA()(),n.TgZ(25,"nz-form-item")(26,"nz-form-label",13),n._uU(27,"\u5916\u90e8"),n.qZA(),n.TgZ(28,"nz-form-control",9)(29,"label",14),n._uU(30,"\u5916\u90e8"),n.qZA()()(),n.TgZ(31,"nz-form-item")(32,"nz-form-label",15),n._uU(33,"\u7248\u672c"),n.qZA(),n.TgZ(34,"nz-form-control",2),n._UZ(35,"input",16),n.YNc(36,y,1,1,"ng-template",null,17,n.W1O),n.qZA()(),n.TgZ(38,"nz-form-item")(39,"nz-form-label",18),n._uU(40,"\u547d\u4ee4"),n.qZA(),n.TgZ(41,"nz-form-control",2),n._UZ(42,"input",19),n.YNc(43,Q,1,1,"ng-template",null,20,n.W1O),n.qZA()(),n.TgZ(45,"nz-form-item")(46,"nz-form-label",21),n._uU(47,"\u662f\u5426\u7981\u7528"),n.qZA(),n.TgZ(48,"nz-form-control",9),n._UZ(49,"nz-switch",22),n.qZA()()()),2&t){const a=n.MAs(7),s=n.MAs(14),h=n.MAs(37),on=n.MAs(44);n.Q6J("formGroup",o.group),n.xp6(2),n.Q6J("nzSm",6)("nzXs",24),n.xp6(2),n.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",a),n.xp6(5),n.Q6J("nzSm",6)("nzXs",24),n.xp6(2),n.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",s),n.xp6(5),n.Q6J("nzSm",6)("nzXs",24),n.xp6(2),n.Q6J("nzSm",14)("nzXs",24),n.xp6(3),n.Q6J("nzSm",6)("nzXs",24),n.xp6(2),n.Q6J("nzSm",14)("nzXs",24),n.xp6(3),n.Q6J("nzSm",6)("nzXs",24),n.xp6(2),n.Q6J("nzSm",14)("nzXs",24),n.xp6(4),n.Q6J("nzSm",6)("nzXs",24),n.xp6(2),n.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",h),n.xp6(5),n.Q6J("nzSm",6)("nzXs",24),n.xp6(2),n.Q6J("nzSm",14)("nzXs",24)("nzErrorTip",on),n.xp6(5),n.Q6J("nzSm",6)("nzXs",24),n.xp6(2),n.Q6J("nzSm",14)("nzXs",24)}},dependencies:[d.O5,u.t3,u.SK,m.Lr,m.Nx,m.iK,m.Fd,l._Y,l.Fj,l.JJ,l.JL,l.sg,l.u,T.Zp,v.i,Z.Ie]}),e})();var c=i(235),P=i(6960),_=i(1634),X=i(5227),D=i(8989),U=i(7327),N=i(1918),I=i(558);function M(e,r){if(1&e){const t=n.EpF();n.TgZ(0,"div",6),n.NdJ("click",function(){const s=n.CHM(t).$implicit,h=n.oxw();return n.KtG(h.informate(s))}),n._UZ(1,"app-card",7),n.qZA()}if(2&e){const t=r.$implicit;n.xp6(1),n.Q6J("img","plugin.png")("id",t.id)("name",t.name)("title","\u63d2\u4ef6")}}let Y=(()=>{class e{constructor(t,o,a,s){this.modal=t,this.router=o,this.rs=a,this.msg=s,this.isVisible=!1,this.loading=!0,this.chooseData={},this.datum=[],this.total=1,this.pageSize=20,this.pageIndex=1,this.query={},this.checked=!1,this.indeterminate=!1,this.setOfCheckedId=new Set,this.delResData=[],this.pageSizeChange(this.pageSize)}reload(){this.datum=[],this.load()}close(){this.isVisible=!1}informate(t){this.isVisible=!0,this.rs.get(`plugin/${t.id}`,this.query).subscribe(o=>{this.chooseData=o.data}).add(()=>{this.loading=!1})}load(){this.loading=!0,this.rs.post("plugin/search",this.query).subscribe(t=>{this.datum=t.data||[],this.total=t.total}).add(()=>{this.loading=!1})}enable(t){this.rs.get(`plugin/${t}/enable`).subscribe(o=>{this.reload()})}disable(t){this.rs.get(`plugin/${t}/disable`).subscribe(o=>{this.reload()})}plugin(t,o){switch(t){case 0:this.rs.get(`plugin/${o}/start`).subscribe(a=>{this.load()});break;case 1:this.rs.get(`plugin/${o}/stop`).subscribe(a=>{this.load()});break;case 2:this.rs.get(`plugin/${o}/restart`).subscribe(a=>{this.load()})}}create(){let t="/plugin/create";location.pathname.startsWith("/admin")&&(t="/admin"+t),this.router.navigateByUrl(t)}handleExport(){this.href="/api/plugin/export"}delete(t,o){this.rs.get(`plugin/${t}/delete`).subscribe(a=>{o?o&&(this.delResData.push(a),o===this.delResData.length&&(this.msg.success("\u5220\u9664\u6210\u529f"),this.load())):(this.msg.success("\u5220\u9664\u6210\u529f"),this.datum=this.datum.filter(s=>s.id!==t))})}onQuery(t){(0,S.r)(t,this.query),this.load()}pageIndexChange(t){this.query.skip=t-1,this.load()}pageSizeChange(t){this.query.limit=t,this.load()}search(t){this.query.keyword={name:t},this.query.skip=0,this.load()}cancel(){this.msg.info("\u53d6\u6d88\u64cd\u4f5c")}handleEdit(t){this.isVisible=!1;const a=this.modal.create({nzTitle:t?"\u7f16\u8f91\u63d2\u4ef6":"\u521b\u5efa\u63d2\u4ef6",nzStyle:{top:"20px"},nzContent:g,nzComponentParams:{id:t},nzMaskClosable:!1,nzFooter:[{label:"\u53d6\u6d88",onClick:()=>{a.destroy()}},{label:"\u4fdd\u5b58",type:"primary",onClick:s=>{s.submit().then(()=>{a.destroy(),this.load()},()=>{})}}]})}handleBatchDel(){(0,c.mK)(this)}handleAllChecked(t){(0,c.Yk)(t,this)}handleItemChecked(t,o){(0,c.mp)(t,o,this)}}return e.\u0275fac=function(t){return new(t||e)(n.Y36(P.Sf),n.Y36(p.F0),n.Y36(f.s),n.Y36(C.dD))},e.\u0275cmp=n.Xpm({type:e,selectors:[["app-plugins"]],decls:7,vars:9,consts:[["placeholder","\u8bf7\u8f93\u5165\u540d\u79f0",3,"onSearch"],["uploadApi","/api/plugin/import","downloadApi","/api/plugin/export",3,"showDelBtn","add"],["nz-row","",1,"card",3,"nzGutter"],["nz-col","","class","card-item",3,"click",4,"ngFor","ngForOf"],[3,"isVisible","title","chooseData","edit","delete","close"],["nzShowSizeChanger","",2,"position","fixed","bottom","0","right","10px",3,"nzPageIndex","nzTotal","nzPageSize","nzPageSizeChange","nzPageIndexChange"],["nz-col","",1,"card-item",3,"click"],[3,"img","id","name","title"]],template:function(t,o){1&t&&(n.TgZ(0,"app-toolbar")(1,"app-search-form",0),n.NdJ("onSearch",function(s){return o.search(s)}),n.qZA(),n.TgZ(2,"app-batch-btn",1),n.NdJ("add",function(){return o.handleEdit()}),n.qZA()(),n.TgZ(3,"div",2),n.YNc(4,M,2,4,"div",3),n.qZA(),n.TgZ(5,"app-detail",4),n.NdJ("edit",function(s){return o.handleEdit(s)})("delete",function(s){return o.delete(s)})("close",function(){return o.close()}),n.qZA(),n.TgZ(6,"nz-pagination",5),n.NdJ("nzPageSizeChange",function(s){return o.pageSizeChange(s)})("nzPageIndexChange",function(s){return o.pageIndexChange(s)}),n.qZA()),2&t&&(n.xp6(2),n.Q6J("showDelBtn",!1),n.xp6(1),n.Q6J("nzGutter",16),n.xp6(1),n.Q6J("ngForOf",o.datum),n.xp6(1),n.Q6J("isVisible",o.isVisible)("title","\u63d2\u4ef6")("chooseData",o.chooseData),n.xp6(1),n.Q6J("nzPageIndex",o.pageIndex)("nzTotal",o.total)("nzPageSize",o.pageSize))},dependencies:[d.sg,_.dE,u.t3,u.SK,X.n,D.A,U.J,N.o,I.q],styles:[".card[_ngcontent-%COMP%]{box-sizing:border-box;padding:10px 20px;width:100%}.card[_ngcontent-%COMP%]   .card-item[_ngcontent-%COMP%]{box-sizing:border-box;padding:5px}.card[_ngcontent-%COMP%]   .card-item[_ngcontent-%COMP%]:hover{cursor:pointer}"]}),e})();var B=i(1346);const O=[{path:"",pathMatch:"full",redirectTo:"list"},{path:"list",component:Y},{path:"detail/:id",component:(()=>{class e{}return e.\u0275fac=function(t){return new(t||e)},e.\u0275cmp=n.Xpm({type:e,selectors:[["app-plugin-detail"]],decls:2,vars:0,template:function(t,o){1&t&&(n.TgZ(0,"p"),n._uU(1,"plugin-detail works!"),n.qZA())}}),e})()},{path:"edit/:id",component:g},{path:"create",component:g},{path:"**",component:B.r}];let $=(()=>{class e{}return e.\u0275fac=function(t){return new(t||e)},e.\u0275mod=n.oAB({type:e}),e.\u0275inj=n.cJS({imports:[p.Bz.forChild(O),p.Bz]}),e})();var V=i(8284),R=i(3325),K=i(7423),W=i(1971),q=i(7096),G=i(6616),L=i(9166),H=i(6903),k=i(269),j=i(2577),w=i(6497),nn=i(6672),tn=i(8521);let en=(()=>{class e{}return e.\u0275fac=function(t){return new(t||e)},e.\u0275mod=n.oAB({type:e}),e.\u0275inj=n.cJS({imports:[d.ez,$,V.wm,R.ip,K.PV,W.vh,_.uK,P.Qp,m.U5,w._p,nn.X,l.UX,T.o7,q.Zf,G.sL,L.Y,H.zf,k.HQ,j.S,v.m,tn.aF,Z.Wr]}),e})()}}]);