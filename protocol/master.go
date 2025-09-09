package protocol

import "github.com/god-jason/iot-master/product"

type WriteLinkFunc func(linker, link_id string, data []byte) error

type Master interface {
	OnAttach(devices []byte)
	OnDetach(devices []byte)
	OnData(data []byte)
	OnSync(request *SyncRequest) (*SyncResponse, error)
	OnRead(request *ReadRequest) (*ReadResponse, error)
	OnWrite(request *WriteRequest) (*WriteResponse, error)
	OnAction(request *ActionRequest) (*ActionResponse, error)
}

type MasterManager interface {
	Get(link_id string) Master
	Close(link_id string) error
	Create(linker, link_id string, options []byte, writer WriteLinkFunc) (Master, error)
	Model(product_id string, model *product.ProductModel) //下发物模型配置
}
