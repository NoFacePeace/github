package file

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/NoFacePeace/github/repositories/go/utils/io"
)

func PersistMetaData(fileName string, data []byte) error {
	tmpFileName := fmt.Sprintf("%s.%d.tmp", fileName, rand.Int())
	f, err := os.OpenFile(tmpFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("os open file error: [%w]", err)
	}
	if _, err := f.Write(data); err != nil {
		return io.SafeClose(f, fmt.Errorf("os file write error: [%w]", err))
	}
	if err := f.Sync(); err != nil {
		return io.SafeClose(f, fmt.Errorf("os file sync error: [%w]", err))
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("os file close error: [%w]", err)
	}
	if err := os.Rename(tmpFileName, fileName); err != nil {
		return fmt.Errorf("os rename error: [%w]", err)
	}
	return nil
}
