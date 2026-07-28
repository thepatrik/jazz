#!/bin/bash

BINARY="$(dirname "$0")/bin/main.exe"
TMPFILE=$(mktemp /tmp/jazz_XXXXXX.jz)
PASS=0
FAIL=0

check() {
    local desc="$1"
    local source="$2"
    local expected="$3"
    printf '%s' "$source" > "$TMPFILE"
    local actual
    actual=$("$BINARY" "$TMPFILE" 2>/dev/null)
    if [ "$actual" = "$expected" ]; then
        printf 'PASS: %s\n' "$desc"
        PASS=$((PASS + 1))
    else
        printf 'FAIL: %s\n      expected: [%s]\n      got:      [%s]\n' "$desc" "$expected" "$actual"
        FAIL=$((FAIL + 1))
    fi
}

check_error() {
    local desc="$1"
    local source="$2"
    local pattern="$3"
    printf '%s' "$source" > "$TMPFILE"
    local actual
    actual=$("$BINARY" "$TMPFILE" 2>&1 1>/dev/null)
    if printf '%s' "$actual" | grep -q "$pattern"; then
        printf 'PASS: %s\n' "$desc"
        PASS=$((PASS + 1))
    else
        printf 'FAIL: %s\n      expected error matching: [%s]\n      got: [%s]\n' "$desc" "$pattern" "$actual"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Arithmetic ==="
check "addition"             "print 1 + 2;"             "3"
check "subtraction"          "print 10 - 4;"            "6"
check "multiplication"       "print 3 * 4;"             "12"
check "division"             "print 10 / 4;"            "2.5"
check "precedence"           "print 1 + 2 * 3 / 4;"    "2.5"
check "grouping"             "print (1 + 2) * 3;"       "9"
check "unary negate"         "print -3;"                "-3"
check "double negate"        "print --3;"               "3"

echo ""
echo "=== Booleans ==="
check "true"                 "print true;"              "true"
check "false"                "print false;"             "false"
check "not true"             "print !true;"             "false"
check "not false"            "print !false;"            "true"

echo ""
echo "=== Nil ==="
check "nil"                  "print nil;"               "nil"
check "not nil"              "print !nil;"              "true"

echo ""
echo "=== Equality ==="
check "1 == 1"               "print 1 == 1;"            "true"
check "1 == 2"               "print 1 == 2;"            "false"
check "1 != 2"               "print 1 != 2;"            "true"
check "1 != 1"               "print 1 != 1;"            "false"
check "true == true"         "print true == true;"      "true"
check "true == false"        "print true == false;"     "false"
check "nil == nil"           "print nil == nil;"        "true"
check "nil == false"         "print nil == false;"      "false"

echo ""
echo "=== Comparison ==="
check "3 > 2"                "print 3 > 2;"             "true"
check "2 > 3"                "print 2 > 3;"             "false"
check "3 >= 3"               "print 3 >= 3;"            "true"
check "2 >= 3"               "print 2 >= 3;"            "false"
check "2 < 3"                "print 2 < 3;"             "true"
check "3 < 2"                "print 3 < 2;"             "false"
check "2 <= 2"               "print 2 <= 2;"            "true"
check "3 <= 2"               "print 3 <= 2;"            "false"

echo ""
echo "=== Variables ==="
check "declaration + read" \
$'let x = 5;\nprint x;' \
"5"

check "uninitialized is nil" \
$'let x;\nprint x;' \
"nil"

check "multiple variables" \
$'let a = 3;\nlet b = 4;\nprint a + b;' \
"7"

check "assignment" \
$'let x = 1;\nx = 99;\nprint x;' \
"99"

check "assign updates value" \
$'let x = 10;\nx = x + 5;\nprint x;' \
"15"

check "multiple print statements" \
$'let a = 1;\nlet b = 2;\nprint a;\nprint b;\nprint a + b;' \
$'1\n2\n3'

echo ""
echo "=== Local variables ==="
check "local declaration + read" \
$'{ let x = 42; print x; }' \
"42"

check "local is nil when uninitialized" \
$'{ let x; print x; }' \
"nil"

check "local assignment" \
$'{ let x = 1; x = 99; print x; }' \
"99"

check "local shadows global" \
$'let x = 1;\n{ let x = 2; print x; }\nprint x;' \
$'2\n1'

check "multiple locals" \
$'{ let a = 3; let b = 4; print a + b; }' \
"7"

check "nested scopes" \
$'{ let a = 1; { let b = 2; print a + b; } }' \
"3"

check "local not visible after scope" \
$'{ let x = 5; }\nprint x;' \
""

check_error "use variable in own initializer" \
$'{ let x = x; }' \
"own initializer"

check_error "duplicate local in same scope" \
$'{ let x = 1; let x = 2; }' \
"Already a variable"

echo ""
echo "=== If / else ==="
check "if true"  $'if (true) print 1;'              "1"
check "if false" $'if (false) print 1;'             ""
check "if else true"  $'if (true) print 1; else print 2;'  "1"
check "if else false" $'if (false) print 1; else print 2;' "2"
check "if condition expression" \
$'let x = 5;\nif (x > 3) print 1; else print 0;' \
"1"
check "nested if" \
$'if (true) { if (false) print 1; else print 2; }' \
"2"

echo ""
echo "=== While ==="
check "while basic" \
$'let i = 0;\nwhile (i < 3) { print i; i = i + 1; }' \
$'0\n1\n2'

check "while false body skipped" \
$'while (false) print 1;' \
""

check "while countdown" \
$'let n = 3;\nwhile (n > 0) { n = n - 1; }\nprint n;' \
"0"

echo ""
echo "=== For ==="
check "for basic" \
$'for (let i = 0; i < 3; i = i + 1) print i;' \
$'0\n1\n2'

check "for sum" \
$'let s = 0;\nfor (let i = 1; i <= 4; i = i + 1) s = s + i;\nprint s;' \
"10"

check "for no initializer" \
$'let i = 0;\nfor (; i < 2; i = i + 1) print i;' \
$'0\n1'

check "for no increment" \
$'for (let i = 0; i < 3;) { print i; i = i + 1; }' \
$'0\n1\n2'

echo ""
echo "=== Strings ==="
check "print string"          'print "hello";'                    "hello"
check "empty string"          'print "";'                          ""
check "concatenation"         'print "hello" + " world";'         "hello world"
check "concat with variable"  'let s = "foo"; print s + "bar";'   "foobar"
check "multi concat"          'print "a" + "b" + "c";'            "abc"
check "string equality"       'print "abc" == "abc";'             "true"
check "string inequality"     'print "abc" == "xyz";'             "false"
check "string != operator"    'print "a" != "b";'                 "true"
check "string in variable"    'let s = "jazz"; print s;'          "jazz"
check "string as argument" \
$'fn greet(name) { return "hello " + name; }\nprint greet("world");' \
"hello world"
check "string + number coercion"   $'print "x" + 1;'    "x1"
check "number + string coercion"   $'print 1 + "x";'    "1x"
check "string + bool coercion"     $'print "v=" + true;' "v=true"
check "string + nil coercion"      $'print "v=" + nil;'  "v=nil"

echo ""
echo "=== And / Or ==="
check "true and true"   "print true and true;"    "true"
check "true and false"  "print true and false;"   "false"
check "false and true"  "print false and true;"   "false"
check "false and false" "print false and false;"  "false"

check "true or true"    "print true or true;"     "true"
check "true or false"   "print true or false;"    "true"
check "false or true"   "print false or true;"    "true"
check "false or false"  "print false or false;"   "false"

check "and short-circuits" \
$'let x = 0;\nfalse and (x = 1);\nprint x;' \
"0"

check "or short-circuits" \
$'let x = 0;\ntrue or (x = 1);\nprint x;' \
"0"

check "and returns left when falsy" \
$'print nil and 2;' \
"nil"

check "or returns left when truthy" \
$'print 1 or 2;' \
"1"

check "chained and" \
$'print true and true and true;' \
"true"

check "chained or" \
$'print false or false or true;' \
"true"

echo ""
echo "=== Functions ==="
check "basic call" \
$'fn greet() { print 42; }\ngreet();' \
"42"

check "return value" \
$'fn add(a, b) { return a + b; }\nprint add(3, 4);' \
"7"

check "implicit nil return" \
$'fn nothing() {}\nprint nothing();' \
"nil"

check "multiple parameters" \
$'fn mul(a, b) { return a * b; }\nprint mul(6, 7);' \
"42"

check "local variables in function" \
$'fn f() { let x = 10; let y = 20; return x + y; }\nprint f();' \
"30"

check "call multiple times" \
$'fn inc(n) { return n + 1; }\nprint inc(0);\nprint inc(5);\nprint inc(9);' \
$'1\n6\n10'

check "recursion" \
$'fn fact(n) { if (n <= 1) return 1; return n * fact(n - 1); }\nprint fact(5);' \
"120"

check "function assigned to variable" \
$'fn double(x) { return x * 2; }\nlet f = double;\nprint f(21);' \
"42"

check "function print value" \
$'fn id(x) { return x; }\nprint id(true);' \
"true"

echo ""
echo "=== Closures ==="
check "counter closure" \
$'fn makeCounter() {\n  let n = 0;\n  fn count() { n = n + 1; return n; }\n  return count;\n}\nlet c = makeCounter();\nprint c(); print c(); print c();' \
$'1\n2\n3'

check "closure captures parameter" \
$'fn adder(x) { fn add(y) { return x + y; } return add; }\nlet add5 = adder(5);\nprint add5(3);\nprint add5(10);' \
$'8\n15'

check "independent closures don'\''t share state" \
$'fn makeCounter() {\n  let n = 0;\n  fn count() { n = n + 1; return n; }\n  return count;\n}\nlet a = makeCounter(); let b = makeCounter();\nprint a(); print a(); print b();' \
$'1\n2\n1'

check "shared upvalue - two closures same var" \
$'fn make() {\n  let n = 0;\n  fn inc() { n = n + 1; }\n  fn get() { return n; }\n  inc(); inc(); inc();\n  return get;\n}\nprint make()();' \
"3"

check "closure mutates captured var" \
$'fn f() {\n  let x = 1;\n  fn set(v) { x = v; }\n  fn get() { return x; }\n  set(42);\n  return get;\n}\nprint f()();' \
"42"

check "nested closure (upvalue of upvalue)" \
$'fn outer(x) {\n  fn middle(y) {\n    fn inner(z) { return x + y + z; }\n    return inner;\n  }\n  return middle;\n}\nprint outer(1)(2)(3);' \
"6"

check "closure over loop variable" \
$'fn makeAdder(n) { fn add(x) { return n + x; } return add; }\nprint makeAdder(10)(1);\nprint makeAdder(20)(1);' \
$'11\n21'

check "returned closure survives enclosing scope" \
$'fn f() { let x = 99; fn g() { return x; } return g; }\nlet g = f();\nprint g();' \
"99"

echo ""
echo "=== Block comments ==="
check "block comment ignored"      $'/* hello */ print 1;'                "1"
check "block comment multiline"    $'/*\n  skip\n*/ print 2;'             "2"
check "inline block comment"       $'print /* skip */ 3;'                 "3"
check "block comment in expr"      $'print 1 /* + 999 */ + 1;'            "2"

echo ""
echo "=== Native: clock ==="
check "clock returns number" \
$'let t = clock(); print t > 0;' \
"true"

check "clock measures elapsed time" \
$'let a = clock(); let b = clock(); print b >= a;' \
"true"

echo ""
echo "=== Runtime errors ==="
check_error "type error: bool + number"  "print true + 1;"   "Runtime error"
check_error "division by zero"           "print 10 / 0;"     "Runtime error"
check_error "negate non-number"          "print -true;"      "Runtime error"
check_error "undefined variable"         "print x;"          "Undefined variable"
check_error "assign undefined variable"  "x = 5;"            "Undefined variable"

echo ""
echo "=== Arrays ==="
check "array literal and index"   $'let a = [1, 2, 3]; print a[0]; print a[1]; print a[2];' $'1\n2\n3'
check "print array"               $'print [1, 2, 3];'                                        "[1, 2, 3]"
check "empty array"               $'let a = []; print len(a);'                               "0"
check "array set element"         $'let a = [1, 2, 3]; a[1] = 99; print a[1];'               "99"
check "array of strings"          $'let a = ["x", "y"]; print a[0] + a[1];'                 "xy"
check "array mixed types"         $'let a = [1, true, "hi"]; print a[2];'                    "hi"
check "array in variable"         $'let a = [10, 20]; let b = a; print b[0];'                "10"
check "array as function arg"     $'fn first(a) { return a[0]; }\nprint first([42, 99]);'    "42"
check "nested arrays"             $'let a = [[1, 2], [3, 4]]; print a[0][1];'                "2"
check "array loop"                $'let a = [0, 0, 0]; for (let i = 0; i < 3; i = i + 1) { a[i] = i * 2; } print a[0]; print a[1]; print a[2];' $'0\n2\n4'

echo ""
echo "=== len / push natives ==="
check "len array"                 $'print len([1, 2, 3]);'                                   "3"
check "len string"                $'print len("hello");'                                     "5"
check "len empty"                 $'print len([]);'                                          "0"
check "push appends element"      $'let a = [1, 2]; push(a, 3); print len(a); print a[2];'  $'3\n3'
check "push returns new length"   $'let a = []; print push(a, "x"); print push(a, "y");'    $'1\n2'

echo ""
echo "=== Array runtime errors ==="
check_error "index out of bounds" $'let a = [1, 2]; print a[5];'                             "out of bounds"
check_error "negative index"      $'let a = [1, 2]; print a[-1];'                            "out of bounds"
check_error "index non-array"     $'let x = 5; print x[0];'                                  "Can only index"
check_error "set out of bounds"   $'let a = [1]; a[5] = 99;'                                 "out of bounds"

echo ""
echo "=== GC: string allocation pressure ==="
check "GC: string concat loop" \
$'let s = "";\nfor (let i = 0; i < 10; i = i + 1) { s = s + "a"; }\nprint s;' \
"aaaaaaaaaa"

check "GC: number coercion concat loop" \
$'let s = "";\nfor (let i = 0; i < 5; i = i + 1) { s = s + true; }\nprint s;' \
"truetruetruetruetrue"

echo ""
echo "=== GC: closures survive collection ==="
check "GC: closure upvalue survives pressure" \
$'fn makeCounter() {\n  let n = 0;\n  fn count() { n = n + 1; return n; }\n  return count;\n}\nlet c = makeCounter();\nlet s = "";\nfor (let i = 0; i < 50; i = i + 1) { s = s + "x"; }\nprint c();\nprint c();' \
$'1\n2'

check "GC: returned closure not collected" \
$'fn f() { let x = 42; fn g() { return x; } return g; }\nlet g = f();\nlet s = "";\nfor (let i = 0; i < 50; i = i + 1) { s = s + "x"; }\nprint g();' \
"42"

echo ""
echo "=== GC: recursion under pressure ==="
check "GC: recursion with allocation" \
$'fn sum(n) { if (n <= 0) return 0; let s = "" + n; return n + sum(n - 1); }\nprint sum(20);' \
"210"

echo ""
echo "=== break ==="
check "break exits while"           $'let i = 0; while (i < 10) { if (i == 3) break; i = i + 1; } print i;'  "3"
check "break exits for"             $'let r = 0; for (let i = 0; i < 10; i = i + 1) { if (i == 5) { r = i; break; } } print r;'  "5"
check "break inner loop only"       $'let s = 0; for (let i = 0; i < 3; i = i + 1) { for (let j = 0; j < 3; j = j + 1) { if (j == 1) break; s = s + 1; } } print s;'  "3"
check "break with locals"           $'let r = 0; for (let i = 0; i < 5; i = i + 1) { let x = i * 2; if (x > 4) break; r = x; } print r;'  "4"
check "break infinite while"        $'let i = 0; while (true) { i = i + 1; if (i == 7) break; } print i;'  "7"
check "multiple breaks one fires"   $'let i = 0; while (i < 10) { i = i + 1; if (i == 3) break; if (i == 7) break; } print i;'  "3"

echo ""
echo "=== continue ==="
check "continue skips rest while"   $'let s = 0; let i = 0; while (i < 5) { i = i + 1; if (i == 3) continue; s = s + i; } print s;'  "12"
check "continue skips rest for"     $'let s = 0; for (let i = 0; i < 5; i = i + 1) { if (i == 2) continue; s = s + i; } print s;'  "8"
check "continue runs increment"     $'let s = 0; for (let i = 0; i < 5; i = i + 1) { if (i == 2) continue; s = s + 1; } print s;'  "4"
check "continue inner loop only"    $'let s = 0; for (let i = 0; i < 3; i = i + 1) { for (let j = 0; j < 3; j = j + 1) { if (j == 1) continue; s = s + 1; } } print s;'  "6"
check "continue with locals"        $'let s = 0; for (let i = 0; i < 5; i = i + 1) { let x = i * 2; if (x == 4) continue; s = s + x; } print s;'  "16"

echo ""
echo "=== break/continue error cases ==="
check_error "break outside loop"    $'break;'                                                  "outside of loop"
check_error "continue outside loop" $'continue;'                                               "outside of loop"
check_error "break in function"     $'fn f() { break; } while (true) { f(); break; }'         "outside of loop"
check_error "continue in function"  $'fn f() { continue; } while (true) { f(); break; }'      "outside of loop"

echo ""
echo "=== GC: native survives pressure ==="
check "GC: clock usable after alloc pressure" \
$'let t = clock();\nlet s = "";\nfor (let i = 0; i < 100; i = i + 1) { s = s + "x"; }\nprint t > 0;' \
"true"

rm -f "$TMPFILE"

echo ""
if [ "$FAIL" -eq 0 ]; then
    printf 'All %d tests passed.\n' "$PASS"
else
    printf '%d passed, %d FAILED.\n' "$PASS" "$FAIL"
    exit 1
fi
