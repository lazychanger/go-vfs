package memory

import (
	"fmt"
	"github.com/lazychanger/filesystem/tests"
	"testing"
)

func TestMemFs(t *testing.T) {
	tests.TestDriver(t, fmt.Sprintf("memory:///?maxsize=%d", 2>>10))
}
