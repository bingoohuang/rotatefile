# go-disk-usage

Get disk usage information like how much space is available, free, and used.

## Compatibility

This works for Windows, MacOS, and Linux although there may some minor variability between what this library reports and
what you get from `df`.

## Usage

```go
import "github.com/bingoohuang/rotatefile/du"
usage := du.New("/path/to")
```
