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
echo "=== Runtime errors ==="
check_error "type error: bool + number"  "print true + 1;"   "Runtime error"
check_error "division by zero"           "print 10 / 0;"     "Runtime error"
check_error "negate non-number"          "print -true;"      "Runtime error"
check_error "undefined variable"         "print x;"          "Undefined variable"
check_error "assign undefined variable"  "x = 5;"            "Undefined variable"

rm -f "$TMPFILE"

echo ""
if [ "$FAIL" -eq 0 ]; then
    printf 'All %d tests passed.\n' "$PASS"
else
    printf '%d passed, %d FAILED.\n' "$PASS" "$FAIL"
    exit 1
fi
