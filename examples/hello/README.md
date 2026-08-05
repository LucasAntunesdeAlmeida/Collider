# Hello, collision

![demo](demo.gif)

The smallest complete Collider program: two objects, one collision event,
one click event. This is the code shown in the main README.

```bash
cd examples/hello
go run .
```

Arrow keys move the blue square. Touching the red box prints "hit!" once
per contact (collision events fire on enter, not every frame). Clicking
the red box destroys it.
