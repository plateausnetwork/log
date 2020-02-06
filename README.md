# Rhizom Log

A simple package for printing, rotate and purge logs

## Building from source

To build from source, you will need the following prerequisites:

- Go 1.13 or greater;
- Git

### Downloading the code

First, clone the project:

```bash
git clone git@github.com:rhizomplatform/log.git /your/directory/of/choice/rhizom
cd /your/directory/of/choice/rhizom
```

### Testing

To run the tests, try `go test`.

### Using as a library

```go
import (
  "github.com/rhizomplatform/fs"
  "github.com/rhizomplatform/log"
)

func myFunc() {
  path := fs.Path{"my/directory/path/not/exists/something"}

  logPurgeMinutes := 10
  logRotateMinutes := 5
  
  // loggin setup
  log.Setup(path, "mysufix", logPurgeMinutes, logRotateMinutes)
  defer log.TearDown()
}

```

## License

For more details about our license model, please take a look at the [LICENSE](LICENSE) file.

**2020**, Rhizom Platform.
