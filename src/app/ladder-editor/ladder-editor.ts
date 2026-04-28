import {Component} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {NzInputModule} from 'ng-zorro-antd/input';
import {lexer} from '../lib/plc/lexer';
import {parser} from '../lib/plc/parser';
import {genLua} from '../lib/plc/lua';
import {genJs} from '../lib/plc/js';

@Component({
  standalone: true,
  selector: 'app-ladder-editor',
  imports: [
    FormsModule, NzInputModule
  ],
  templateUrl: './ladder-editor.html',
  styleUrl: './ladder-editor.scss',
})
export class LadderEditor {
  st = `
(*
 这是
 多行
 注释2
*)

 VAR
    a : INT := 0;
    b : INT := 0;
    c : INT := 0;
    flag : BOOL := FALSE;
    counter : INT;

    TON1 : TON;
    TOF1  : TOF
    TP1 : TP
END_VAR


a := 10;
b := 20;
c := a + b * 2;

// 这是注释3
flag := TRUE;

IF a > b THEN
    c := 1;
ELSIF a = b THEN
    c := 2;
ELSE
    c := 3;
END_IF;

CASE a OF
    0:
        c := 100;
    10:
        c := 200;
    20:
        c := 300;
END_CASE;

FOR counter := 0 TO 5 DO
    c := c + counter;
END_FOR;

WHILE counter < 10 DO
    counter := counter + 1;
END_WHILE;

c := add(a, b);

TON1(IN := flag, PT := T#5s);
TOF1(IN := flag, PT := T#3s);
TP1(IN := flag, PT := T#1s);
  `
  lua = ""

  convert() {
    let ts = lexer(this.st)
    console.log("tokens", ts)
    let ast = parser(ts)
    console.log("AST", ast)
    this.lua = genLua(ast)
  }

  convertJS() {
    let ts = lexer(this.st)
    console.log("tokens", ts)
    let ast = parser(ts)
    console.log("AST", ast)
    this.lua = genJs(ast)
  }
}
