package main

import (
	"errors"
	"github.com/kardianos/service"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	var serviceConfig = &service.Config{
		Name:        "iot-master",
		DisplayName: "物联大师",
		Description: "物联大师边缘计算网关",
		Arguments:   nil,
	}

	p := &Program{}

	svc, err := service.New(p, serviceConfig)
	if err != nil {
		log.Fatal(err)
	}

	status, err := svc.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			err = svc.Install()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	log.Println("status", status)

	err = svc.Run()
	if err != nil {
		log.Fatal(err)
	}
}

type Program struct {
}

func (p *Program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *Program) Stop(s service.Service) error {

	return nil
}

func (p *Program) run() {
	//
	//// 此处编写具体的服务代码
	hup := make(chan os.Signal, 2)
	signal.Notify(hup, syscall.SIGHUP)
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, os.Kill)

	//cmd := exec.Command("iot-master")
	proc, err := os.StartProcess("iot-master", nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	exit := false

	go func() {
		for {
			select {
			case <-hup:
			case <-quit:
				//_ = p.Shutdown() //全关闭两次
				exit = true
				_ = proc.Kill()
				os.Exit(0)
			}
		}
	}()

	//循环等待
	for {
		state, err := proc.Wait()
		if err != nil {
			log.Println(err)
		} else {
			log.Println(state.String())
		}

		if exit {
			break
		}

		time.Sleep(time.Second * 5) //默认5s重启

		//重来
		proc, err = os.StartProcess("iot-master", nil, nil)
		if err != nil {
			log.Fatal(err)
		}
	}

}
