$comptime: 24
$comptime: console.log("any expression can be used in comptime, it will be expunged in the runtime code.")
$comptime: {
  console.log("multiple lines can be used")
  console.log("with the power of the usually rarely useful javascript scope")
}

function subScope() {
  $comptime: console.log("comptime can be used anywhere.")
}

$comptime: (function() {
  $comptime: console.log("$comptime cannot be used within $comptime, I don't want to implement it, and if you need this, you're probably abusing comptime.")
})()
