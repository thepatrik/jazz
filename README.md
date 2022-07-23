# Jazz

Jazz is a small programming language created to learn about... programming languages.

To launch `jazz` REPL.

```console
$ make jazz
```

### Syntax

The syntax looks something like this.

```
fn fib(n) {
    if (n <= 1) return n;
    return fib(n-2) + fib(n-1);
}

for (let i = 0; i < 20; i = i + 1) {
    let num = fib(i);
    print num;
}
```