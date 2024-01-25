$comptime: import { readFileSync as comptimeReadFileSync } from "fs"

$comptime: function comptimeFibonacci(n) {
  let a = 0
  let b = 1
  for (let i = 0; i < n; i++) {
    const newA = b
    b = a + b
    a = newA
  }
  return a
}

$comptime: const unlabeledVariable = 48484848

console.log(
  "this is an inlined variable + the 32nd fibonacci number",
  unlabeledVariable + comptimeFibonacci(32),
)
console.log("README.md", comptimeReadFileSync("README.md", "utf8"))
