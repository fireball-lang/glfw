# GLFW
Automatically generated [Fireball](https://github.com/fireball-lang/fireball) bindings for [GLFW 3.5.1](https://github.com/glfw/glfw/tree/3.5.1).

## Features
- Documentation comments are fully preserved
- Structs names are converted to pascal case and the `GLFW` prefix is stripped (e.g. `GLFWwindow` -> `Window`)
- Function names are converted to snake case and the `glfw` prefix is stripped (e.g. `glfwCreateWindow` -> `create_window`)
- Macro constants are converted to enums (e.g. `GLFW_KEY_A` -> `Key::A`)
- If possible, functions that take in (or return) an `int` "enum" actually take in the strongly typed enum (e.g. `get_key_name(key: i32)` -> `get_key_name(key: Key)`)
- Structs have method wrappers with the type name removed (e.g. it's possible to call `should_close()` on an instance of `Window`)
- `Window` and `Cursor` structs have manually written `new` static methods that return a "safe" optional reference
