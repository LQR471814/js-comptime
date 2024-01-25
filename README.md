## js-comptime

> A transpiler which adds "comptime" aka. "compile-time evaluation" aka. "partial evaluation" to javascript.

Examples of how it should be used can be found under [examples](examples/)

### Tour

The primary addition the `js-comptime` compiler adds to javascript is the `$comptime` label. "comptime" is essentially the execution of code at compile-time instead of runtime.

You can add the `$comptime` [label](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/label) before most javascript statements to have them execute at compile time.

```js
// before build
$comptime: console.log("this debug line will only be shown when comptime compiles the application")
console.log("hello world")
```

```js
// after build
console.log("hello world")
```

You can reference variables and functions defined in `$comptime` from "runtime" code and the value they evaluate to will be inlined.

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

### Implementation

1. For each scope.
1. Detect all "comptime values", variable or function declarations preceded by the `$comptime:` label.
1. Detect all "comptime regions", runtime operations/expressions which only rely on comptime values (from current or parent scopes) or constants.
1. Execute "comptime values" and "comptime regions" while inlining the result of "comptime regions" in their appropriate areas in order from top to bottom.

### Configuration

- Entrypoint(s) can be configured.
- Inlining code in `node_modules` is disabled by default. A whitelist and blacklist with glob support can be provided.
- The runtime in which `$comptime` code is executed can be configured. (ex. nodejs, browser window, etc...)