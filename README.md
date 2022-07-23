# Jazz

Jazz is a small programming language created to learn about... programming languages. Based on the book *Crafting Interpreters*, by Robert Nystrom.

The syntax looks something like this.

```rust
fn fib(n) {
    if (n <= 1) return n;
    return fib(n-2) + fib(n-1);
}

for (let i = 0; i < 20; i = i + 1) {
    let num = fib(i);
    print num;
}
```

To launch the `jazz` REPL.

```console
$ make jazz
Welcome to Jazz v0.0.1
Type ".exit" to exit.
> 1+2*3/4;
2.5
```