import {Component} from '@angular/core';
import {st2lua} from './st2lua'
import {FormsModule} from '@angular/forms';
import {NzInputModule} from 'ng-zorro-antd/input';
import {lexer} from '../lib/plc/lexer';
import {parser} from '../lib/plc/parser';
import {genLua} from '../lib/plc/lua';

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
 VAR
    a : INT;
    b : INT;
    c : INT;
    flag : BOOL;
    counter : INT;
END_VAR

a := 10;
b := 20;
c := a + b * 2;

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
    this.lua = st2lua(this.st, true)
    let ts = lexer(this.st)
    console.log("tokens", ts)
    let ast = parser(ts)
    console.log("AST", ast)
    this.lua = genLua(ast)
  }
}
