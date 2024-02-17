$comptime: const { readFileSync: comptimeReadFileSync, access } = require("fs")

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
$comptime: const a = 48484848, b = 3
$comptime: let x
$comptime: const [y, z] = [4, 3]

console . log(
  "this is an inlined variable + the 32nd fibonacci number",
  unlabeledVariable + comptimeFibonacci(32),
)
console.log("README.md", comptimeReadFileSync("go.mod", "utf8"))
