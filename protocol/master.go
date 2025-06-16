package protocol

import "github.com/god-jason/iot-master/link"

type WriteLinkFunc func(linker, link_id string, data []byte) error

type Master interface {
	OnData(data []byte)
	OnSync(request *SyncRequest) (*SyncResponse, error)
	OnRead(request *ReadRequest) (*ReadResponse, error)
	OnWrite(request *WriteRequest) (*WriteResponse, error)
	OnAction(request *ActionRequest) (*ActionResponse, error)
	OnAttach(devices []link.Device)
	OnDetach(devices []string)
}

type MasterManager interface {
	Get(link_id string) Master
	Close(link_id string) error
	Create(linker, link_id string, options []byte, writer WriteLinkFunc) (Master, error)
	Config(product_id string, config []byte) //产品协议配置
}
