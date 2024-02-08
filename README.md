## js-comptime

> A transpiler which adds compile-time evaluation, or partial evaluation to JavaScript.

Examples of how it should be used can be found under [examples](examples/)

### Tour

The `js-comptime` compiler adds metaprogramming capabilities to javascript via the `$comptime` label.

#### Code execution

You can add the `$comptime` [label](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/label) before a block or an expression to have it only execute during compile time.

```js
// before build
$comptime: console.log("this debug line will only be shown when comptime compiles the application")
console.log("hello world")
```

```
// during build
this debug line will only be shown when comptime compiles the application
```

```js
// after build
console.log("hello world")
```

#### Variable declaration

If a variable, function or class is defined right after the `$comptime` label, they can be referenced from runtime code and they'll be inlined with the value they evaluate to during compile time.

```js
// before build
$comptime: const sum = 700 + 30
console.log(sum)

$comptime: function square(n) {
  return n ** 2
}
console.log(square(8))
```

```js
// after build
console.log(730)
console.log(64)
```

Strictly speaking, all operations (number arithmetic/boolean arithmetic, function calls, property access, etc...) which depend on only constants and `$comptime` values will be evaluated at compile time.

```js
// before build
$comptime: const someValue = 4 * 3
console.log(someValue + 32)

$comptime: const obj = { a: { b: "foo" } }
console.log(obj.a.b)
```

```js
// after build
console.log(44)
console.log("foo")
```

> The reason why `console.log(...)` isn't also executed at compile time and the build result being an empty file is because one of its dependencies is the runtime function `console.log`. Obviously, the dependencies of a function execution also include the function being executed. Operations like `+` and `-` can be considered constant, but something like `console.log` which pertains to the execution environment cannot be considered constant.

Objects and arrays can also be returned by `$comptime`, however all their properties/elements must also be inlined.

```js
// before build
import { readFileSync } from "fs"

$comptime: somevalue = "hoho"
$comptime: const obj = {
  somevalue: somevalue,
  file: JSON.parse(readFileSync("config.json", "utf8")),
  foo: "bar"
}
$comptime: const arr = [obj]

console.log(obj)
console.log(arr)
```

```js
// after build
console.log({
  somevalue: "hoho",
  file: { ... },
  foo: "bar"
})
console.log([{
  somevalue: "hoho",
  file: { ... },
  foo: "bar"
}])
```

Functions can also be returned by `$comptime`, code inside the function body returned by `$comptime` will be treated as "runtime" code, meaning that `$comptime` variables within the function body will be inlined.

```js
// before build
$comptime: function createDoublePrinter(message) {
  return () => console.log(message + message)
}
const fooPrinter = createDoublePrinter("foobar")
fooPrinter()
```

```js
// after build
const fooPrinter = () => console.log("foobarfoobar")
fooPrinter()
```

#### Conditional generation

Code can also be generated conditionally with `if-statements`, the body of the if statement is treated as runtime code, the expression is a comptime statement that determines if the runtime code will be generated or not. To ensure that variables don't conflict with the outside scope, generated code will always be wrapped in a `{ scope }`.

```js
// before build
$comptime: const verbose = process.env.VERBOSE
$comptime: if (verbose === "true") {
  console.log("verbose logging...")
}
console.log("hello world")
```

```js
// after build VERBOSE=true
{
  console.log("verbose logging...")
}
console.log("hello world")
```

```js
// after build VERBOSE=false
console.log("hello world")
```

Code can also be generated repeatedly with `for/while` loops, similarly, generated code will be wrapped in a `{ scope }`.

```js
// before build
const db = someOrm(...)
$comptime: const tableNames = ["user", "student", "teacher"]
$comptime: for (const table of tableNames) {
  db.delete(table).run()
}
```

```js
// after build
{
  db.delete("user").run()
}
{
  db.delete("student").run()
}
{
  db.delete("teacher").run()
}
```

### Implementation

1. For each scope.
1. Detect all "comptime values", variable or function declarations preceded by the `$comptime:` label.
1. Detect all "comptime regions", runtime operations/expressions which only rely on comptime values (from current or parent scopes) or constants.
1. Execute "comptime values" and "comptime regions" while inlining the result of "comptime regions" in their appropriate areas in order from top to bottom.

### Notes

The list of expressions were taken from [MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Expressions_and_Operators).

| Constant | Example |
| --- | --- |
| Null | `null` |
| undefined | `undefined` |
| Any number literal | `23` or `-42.3` |
| Any string literal | `"a string"` or `'another string'` |
| Boolean | `true` or `false` |

| Expression | Example |
| --- | --- |
| Identifier reference | `variableIdentifier` |
| Assignment | `v = <expr>` or `v++` or `v += <expr>` |
| Array literal | `[1, 2]` |
| Object literal | `{ foo: "bar" }` |
| Function call | `func(arg1, ...)` or `(expr)(arg1, ...)` |
| Binary expression | `<expr> + <expr>` or `<expr> in <expr>` |
| Ternary operator | `<expr> ? <expr> : <expr>` |
| Comma operator | `<expr>, <expr>` |
| Grouping operator | `(<expr>)` |
| Format string | ```hello ${<expr>}``` |
| Function definition | `function name(arg1, arg2) { <expr>; ... }` or `() => { <expr> }` |

### Configuration

- Entrypoint(s) can be configured.
- Inlining code in `node_modules` is disabled by default. A whitelist and blacklist with glob support can be provided.
- The runtime in which `$comptime` code is executed can be configured. (ex. nodejs, browser window, etc...)

### Credits

Ideas of comptime are nothing new, attempts at JavaScript comptime like [vite-plugin-compile-time](https://github.com/egoist/vite-plugin-compile-time) already exist. Various ideas from metaprogramming in other languages (like generics/comptime, code generation, introspection) mixed with an unhealthy dose of JavaScript programming culminated into this thing.

MDN's [article](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Expressions_and_Operators) on expressions was pretty helpful.
