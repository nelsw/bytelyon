package s3

import (
	"testing"

	"github.com/nelsw/bytelyon/pkg/logs"
)

func TestListDirectories(t *testing.T) {
	logs.Init("trace")
	ListDirectories("FireFibers.com/pages/resources-how-to/")
}
