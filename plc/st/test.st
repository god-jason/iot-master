PROGRAM test

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
    TOF1  : TOF;
    TP1 : TP;
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

END_PROGRAM
