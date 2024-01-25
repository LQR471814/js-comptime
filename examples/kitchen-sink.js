$comptime: function expensiveFunction() {
  let text = ""
  for (let i = 0; i < 12; i++) {
    text += "Hello world "
  }
  text += "\n"
  return text
}

$comptime: const unlabeledVariable = 48484848

const runtimeValue = expensiveFunction()

$comptime: {
  console.log("this code should only execute at compile time.")
}

console.log("this is an inlined variable", unlabeledVariable)
console.log("this is a variable computed at compile time", runtimeValue)

function runtimeFunction() {
  $comptime: function expensiveMath(n) {
    return 24**n
  }
  $comptime: const x = 42 + 32
  $comptime: {
    // THIS IS INVALID!
    console.log(runtimeValue)
  }

  function update() {
    $comptime: {
      const someOtherConstant = 43
    }
    return x + expensiveMath(x)
  }

  const runtimeValue = expensiveMath(1)
  return [x + runtimeValue, update()]
}

runtimeFunction()
