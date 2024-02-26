## jscomptime

> A transpiler which adds compile time metaprogramming to javascript.

Examples of how it should be used can be found under [examples](examples/)

### Tour

The `jscomptime` compiler adds metaprogramming capabilities to javascript via the `$comptime` label.

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

> Note: "constant statements" will not be executed at compile time, so something like this would not work. This is because of the complexity that lies within tracking mutations, therefore it's better just not to support it.

```js
// before build
let x = 0
x += 1
x += 2
console.log(x)

// after build
let x = 0
x += 1
x += 2
console.log(x)
```

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

### Code expansion

Conditional statements and loops can be expanded with the `$expand:` label.

```js
// before build
$comptime: const configuration = {
  verboseLogs: true,  
  feature1: {
    enabled: true
  },
  feature2: {
    enabled: true,
    serveOn: ["127.0.0.1", "0.0.0.0"],
  },
}

$expand: if (configuration.verboseLogs) {
  console.log("verbose logging enabled.")
}
window.features = {}
$expand: for (const key in configuration) {
  window.features[key] = configuration[key]
}
```

```js
// after build
{
  console.log("verbose logging enabled.")
}
window.features = {}
{
  window.features["verboseLogs"] = true
}
{
  window.features["feature1"] = { enabled: true }
}
{
  window.features["feature2"] = { enabled: true, serveOn: ["127.0.0.1", "0.0.0.0"] }
}
```

### Implementation

1. For each scope.
1. Detect all "comptime values", variable or function declarations preceded by the `$comptime:` label.
1. Detect all "comptime regions", runtime operations/expressions which only rely on comptime values (from current or parent scopes) or constants.
1. Execute "comptime values" and "comptime regions" while inlining the result of "comptime regions" in their appropriate areas in order from top to bottom.

#### In detail

A "scope" tree is created, each node holding the comptime and runtime variables declared in the scope, the comptime statements within the scope, and child scopes to the current scope.

A child scope is created when the following are encountered:
- A lexical block `{ statement; }`, in this case a new empty scope is created.
- A function declaration `function (arg1, arg2) { ... }`, in this case a new empty scope is created and the function arguments are used as runtime declarations.
- A arrow function `(arg1, arg2) => { ... }`, this is the same as a function declaration.
- A method declaration `name(arg1, arg2) { ... }`, this is the same as a function declaration.

A "variable declaration" is:
- A lexical variable declaration.
- A function declaration.
- A class declaration.

`this` is always treated as a runtime value, trying to assign `this.<something>` within `$comptime` will shoot you in the foot.

A "comptime value" is a "variable declaration" which is labeled with `$comptime`.

A "comptime region" is an expression in which all children are comptime. Children which are exempt from a comptime/runtime label are the following:
- `name: <identifier>`

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
