package serial

import (
	"github.com/god-jason/bucket/boot"
	"github.com/god-jason/bucket/log"
)

func init() {
	boot.Register("serial", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"database"},
	})
}

func Startup() error {

	return LoadSerials()
}

func Shutdown() error {
	serials.Range(func(name string, serial *Serial) bool {
		err := serial.Close()
		if err != nil {
			log.Error(err)
		}
		return true
	})

	return nil
}
