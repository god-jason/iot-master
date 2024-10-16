package serial

import (
	"github.com/zgwit/iot-master/db"
	"xorm.io/xorm"
)

var serials lib.Map[Serial]

func LoadSerials() error {
	var serials []*Serial
	err := db.Engine.Find(&serials)
	if err != nil {
		if err == xorm.ErrNotExist {
			return nil
		}
		return err
	}
	for _, m := range serials {
		if m.Disabled {
			continue
		}
		go func(m *Serial) {
			err := LoadSerial(m)
			if err != nil {
				log.Error(err)
			}
		}(m)
	}
	return nil
}

func LoadSerial(m *Serial) error {
	serials.Store(m.Id, m)
	return m.Open()
}

func GetSerial(id string) *Serial {
	return serials.Load(id)
}
