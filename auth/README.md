
# Go Base CRUD

## Quick Start

### Clean all generated files
```bash
sh dev.sh clean
```

### Full regeneration and run (after model changes)
```bash
sh dev.sh run
```

### Development mode with hot-reload
```bash
sh dev.sh dev
```

### Quick run (without regeneration)
```bash
go run main.go
```

## Alternative: Using Make

```bash
make clean    # Clean generated files
make gen      # Generate code
make run      # Regenerate and run
make dev      # Development with hot-reload
make build    # Build binary
``` 
