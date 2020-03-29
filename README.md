# skimo
A cpp file inliner, replace relative include statements with their actual code.

## Usage
- Skimo can be used as a standalone CLI app, or you can use it as a library in your own projects.

- Imagine having this directory structure:
```
.
├── lib
│   ├── centroid_decomp.h
│   ├── contants.h
│   └── tree.h
└── main.cpp
```

- You are working on the `main.cpp`

```c
// main.cpp
#include <iostream>
#include "centroid_decomp.h"

// centroid_decomp.h
#include <memory.h>
#include "constants.h"
#include "tree.h"

// tree.h
#include <vector>
#include "constants.h"

// constants.h

#include <iostream>

const int INF = 1 << 28;
``` 


- The intended contents of the result file

```c
#include <iostream>
const int INF = 1 << 28;

#include <vector>
#include <memory.h>

// rest of the main.cpp code
```

- This is especially usefull because it will enable to you to have library that you can include without
cluttering the code with copy paste, and when trying to submit your code, the contents will be essentially copied over
to your generated file.

- Usage scenario

```
cat main.cpp | skimo --include_dir=lib
```

```go
import (
    "github.com/chermehdi/skimo"
    "fmt"
)

func main() {
    inlinedContent, err := skimo.Inline(content, "lib")
    if err != nil {
        panic(err)
    }
    fmt.Println(inlinedContent)
}
```