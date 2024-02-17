$comptime: const comptimeVar = 24
$comptime: function comptimeFunction(a, b) {
  return n ** 2
}
$comptime: const comptimeKey = "foo"

comptimeVar;
comptimeVar = 32;
const arrayLiteral = [comptimeVar, 42];
const objectLiteral = { comptimeVar, [comptimeKey]: comptimeVar };
comptimeFunction(42, comptimeVar);
43 + comptimeVar;
comptimeKey in objectLiteral;
comptimeVar % 2 === 0 ? comptimeKey : undefined;
comptimeVar, 32;
(comptimeVar, 32);
`format string ${comptimeVar} ${32}`;
(() => comptimeVar * 3)();
objectLiteral.foo;
objectLiteral[comptimeKey];
arrayLiteral[1];
comptimeVar++;
comptimeVar += 1;
delete objectLiteral[comptimeKey];
comptimeFunction`something ${23, 32, 48}`;
for (let i = 0; i < 23; i++) {}
for (const value of [2, 3, 4]) {}
while (true) {}
if (!false) {}
switch (true) {
  case true:
    break
  case false:
    break
}

class What { what() {} }
