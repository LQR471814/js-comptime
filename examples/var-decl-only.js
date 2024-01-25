$comptime: const unlabeledVariable = 48484848
$comptime: const a = 48484848, b = 3
$comptime: let x
$comptime: const [y, z] = [4, 3]
$comptime: const { foo, bar: { x: [what] } } = { foo: "", bar: { x: [2] } }

$comptime: function sum(a, b) {
  return a + b
}
