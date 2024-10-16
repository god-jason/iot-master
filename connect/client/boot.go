package client

func init() {
	boot.Register("client", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"database"},
	})
}

func Startup() error {

	return LoadClients()
}

func Shutdown() error {
	clients.Range(func(name string, client *Client) bool {
		err := client.Close()
		if err != nil {
			log.Error(err)
		}
		return true
	})

	return nil
}
