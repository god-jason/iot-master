package protocol

type WriteLinkFunc func(linker, link_id string, data []byte) error

type Master interface {
	OnData(data []byte)
}

type MasterManager interface {
	Get(link_id string) Master
	Close(link_id string) error
	Create(linker, link_id string, options []byte, writer WriteLinkFunc) (Master, error)
}
