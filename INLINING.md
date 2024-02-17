## inlining

Where can compile-time code be inlined?
This is the most difficult part of the whole thing.
Here are some considerations.

Different modifications to the source code will be defined as follows:

1. CAN BE INLINED - refers a segment of source code that can be evaluated at compile time, then the result of the evaluation inlined into the source.
1. CANNOT BE INLINED - refers to a segment of source code which must exist after compile time evaluation, all parent source code will also be marked CANNOT BE INLINED.
1. INLINE RESULT IS - refers to what the final inline result is.
1. SEARCH IN - refers to the sub-segments of source code that will be searched for inlineable source code.
1. EXPUNGED - refers to a segment of source code that will be completely removed from the final transpiled code.

### comptime statements

All `$comptime:` statements will be expunged.

### statements

1. An `expression_statement` can be inlined if the expression can be inlined, search in the expression.
1. A `statement_block` can be inlined if everything within the block can be inlined, search in each statement.
1. An `if_statement` can be inlined:
   1. if condition inlineable and if condition true, inline result is block, search in block.
   1. if condition inlineable and if condition false, `if_statement` is expunged.
   1. if condition uninlineable, inline result is if condition, search in condition and block.
1. A `switch_statement` can be inlined with similar logic as `if_statement`.
   1. Search in condition.
   1. Search in each case expression.
   1. If all cases inlineable and condition also inlineable, inline result is corresponding case's block, if none match, `switch_statement` is expunged.
1. A `try_statement` can be inlined if it's block can be inlined, search in body and catch block.
1. No declarations can be inlined:
   1. `class_declaration`, search in each method definition of class.
   1. `function_declaration`, search in function body.
   1. `generator_function_declaration`, search in function body.
   1. `lexical_declaration`, search in value expression.
   1. `variable_declaration`, search in value expression.
1. All other statements cannot be inlined.

### expressions

1. All literals can be inlined.
1. An identifier can be inlined if within expression context and resolves to a comptime variable.
1. A `rest_pattern` can be inlined if it's children can be inlined and within expression context, search within expression.
1. An `x` can be inlined if all if it's children can be inlined, search in children, where `x` can be:
   1. `array`
   1. `template_string`
   1. `ternary_expression`
1. An object can be inlined if all the dynamic keys can be inlined, and all the values can be inlined, search in all dynamic keys and values.
1. Any `binary_expression` expression can be inlined, search in left and right expressions.
1. A `new_expression` expression can be inlined, search in arguments and function.
1. An `await_expression` expression can be inlined, search in expression.
1. A unary expression with operator if expression can be inlined, search in expression:
   1. `typeof`
   1. `+`, `-`, `~`, `!`
   1. `void`
1. A `member_expression` operator if the object can be inlined, search in object.
1. A `subscript_expression` can be inlined if subscript expression and object can be inlined, search in object and subscript expression.
1. A `call_expression` can be inlined if the function can be inlined and so can all the arguments, search in function and argument expressions.
1. A `super` cannot be inlined.
1. An `function_expression`, `arrow_function`, or `generator_function` cannot be inlined, but the body block will be searched in.
1. A `spread_element` can be inlined if within expression context and element can be inlined.
1. A `parenthesized_expression` can be inlined if not:
   1. `if`, `switch`, `while`, `catch`

### everything else

1. Should be searched in.

