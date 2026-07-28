#!/bin/bash

BINARY="$(dirname "$0")/bin/main.exe"
PASS=0
FAIL=0

check() {
    local expr="$1"
    local expected="$2"
    local actual
    actual=$(printf '%s\n' "$expr" | "$BINARY" 2>/dev/null | sed -n '3s/^> //p')
    if [ "$actual" = "$expected" ]; then
        printf 'PASS: %-30s => %s\n' "$expr" "$expected"
        PASS=$((PASS + 1))
    else
        printf 'FAIL: %-30s expected: %s  got: %s\n' "$expr" "$expected" "$actual"
        FAIL=$((FAIL + 1))
    fi
}

check_error() {
    local expr="$1"
    local pattern="$2"
    local actual
    actual=$(printf '%s\n' "$expr" | "$BINARY" 2>&1 1>/dev/null)
    if printf '%s' "$actual" | grep -q "$pattern"; then
        printf 'PASS: %-30s => (error: %s)\n' "$expr" "$pattern"
        PASS=$((PASS + 1))
    else
        printf 'FAIL: %-30s expected error matching: %s  got: %s\n' "$expr" "$pattern" "$actual"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Arithmetic ==="
check "1 + 2"              "3"
check "10 - 4"             "6"
check "3 * 4"              "12"
check "10 / 4"             "2.5"
check "1 + 2 * 3 / 4"     "2.5"
check "(1 + 2) * 3"        "9"
check "-3"                 "-3"
check "--3"                "3"

echo ""
echo "=== Booleans ==="
check "true"               "true"
check "false"              "false"
check "!true"              "false"
check "!false"             "true"

echo ""
echo "=== Nil ==="
check "nil"                "nil"
check "!nil"               "true"

echo ""
echo "=== Equality ==="
check "1 == 1"             "true"
check "1 == 2"             "false"
check "1 != 2"             "true"
check "1 != 1"             "false"
check "true == true"       "true"
check "true == false"      "false"
check "nil == nil"         "true"
check "nil == false"       "false"

echo ""
echo "=== Comparison ==="
check "3 > 2"              "true"
check "2 > 3"              "false"
check "3 >= 3"             "true"
check "2 >= 3"             "false"
check "2 < 3"              "true"
check "3 < 2"              "false"
check "2 <= 2"             "true"
check "3 <= 2"             "false"

echo ""
echo "=== Runtime errors ==="
check_error "true + 1"     "Runtime error"
check_error "10 / 0"       "Runtime error"
check_error "-true"        "Runtime error"

echo ""
if [ "$FAIL" -eq 0 ]; then
    printf 'All %d tests passed.\n' "$PASS"
else
    printf '%d passed, %d FAILED.\n' "$PASS" "$FAIL"
    exit 1
fi
