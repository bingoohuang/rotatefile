package du

import (
	"fmt"
	"testing"
)

const KB = 1024

func TestNewDiskUsage(t *testing.T) {
	usage, _ := NewDiskUsage(".")
	fmt.Println("Free:", usage.Free()/(KB*KB))
	fmt.Println("Available:", usage.Available()/(KB*KB))
	fmt.Println("Size:", usage.Size()/(KB*KB))
	fmt.Println("Used:", usage.Used()/(KB*KB))
	fmt.Println("Usage:", usage.Usage()*100, "%")
}
