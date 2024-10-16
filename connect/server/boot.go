package server

func init() {
	boot.Register("server", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"database"},
	})
}

func Startup() error {

	return LoadServers()
}

func Shutdown() error {
	servers.Range(func(name string, server *Server) bool {
		err := server.Close()
		if err != nil {
			log.Error(err)
		}
		return true
	})

	return nil
}
