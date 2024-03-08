package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"lion/pkg/guacd"
)

var (
	cfgFile = flag.String("f", "config.json", "config json file path")
	num     = flag.Int("n", 1, "number of connections")
)

func parseConfig(path string) (cfg map[string]interface{}, err error) {
	// parse config file
	p, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("parse config file failed: %s ", err)
	}
	// parse config
	err = json.Unmarshal(p, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("parse config file failed: %s ", err)
	}
	return cfg, nil
}

func main() {
	flag.Parse()
	cfg, err := parseConfig(*cfgFile)
	if err != nil {
		log.Fatalf("parse config file failed: %s ", err)
	}
	log.Printf("config: %+v\n", cfg)
	var addrs []interface{}
	guacamoleAddress := cfg["guacamole_address"]
	if guacamoleAddress != nil {
		addrs = guacamoleAddress.([]interface{})
		log.Printf("guacamole_address is %v\n", addrs)
	}
	guacCfg := guacd.NewConfiguration()
	for k, v := range cfg {
		switch ret := v.(type) {
		case string:
			guacCfg.SetParameter(k, ret)
		case int, int32, int64, uint, uint32, uint64,
			float32, float64:
			guacCfg.SetParameter(k, fmt.Sprintf("%v", ret))
		default:
			log.Printf("unsupported type: %T %s\n", ret, v)
		}
	}
	guacCfg.Protocol = guacCfg.GetParameter("protocol")
	guacdHost := guacCfg.GetParameter("guacamole_host")
	guacdPort := guacCfg.GetParameter("guacamole_port")
	info := guacd.NewClientInformation()
	tunnelCons := &sync.Map{}
	count := 0
	go waitSignal(tunnelCons)
	log.Printf("guacdHost: %s, guacdPort: %s\n", guacdHost, guacdPort)
	addr := net.JoinHostPort(guacdHost, guacdPort)
	var wg sync.WaitGroup
	for i := 0; i < *num; i++ {
		selectAddr := randomAddr(addrs, addr)
		tunnel, err1 := guacd.NewTunnel(selectAddr, guacCfg, info)
		if err1 != nil {
			log.Printf("new tunnel failed: %s\n", err1)
			continue
		}
		wg.Add(1)
		count++
		tunnelCons.Store(tunnel.UUID, tunnel)
		log.Printf("New connection uuid: %+v\n", tunnel.UUID)
		go func(tunnel *guacd.Tunnel) {
			defer wg.Done()
			handle(tunnel)
			tunnelCons.Delete(tunnel.UUID)
			log.Printf("connection %s leave\n", tunnel.UUID)
		}(tunnel)
		time.Sleep(time.Second)
	}
	log.Printf("waiting for all connections %d to close\n", count)
	wg.Wait()
	log.Printf("all connections closed\n")

}

func handle(tunnel *guacd.Tunnel) {
	defer tunnel.Close()
	syncChan := make(chan *guacd.Instruction, 10)
	go func() {
		defer func() {
			close(syncChan)
		}()
		for {
			inst, err1 := tunnel.ReadInstruction()
			if err1 != nil {
				log.Printf("%s read instruction failed: %s\n", tunnel.UUID, err1)
				return
			}

			switch inst.Opcode {
			case guacd.InstructionClientNop:
			case guacd.InstructionClientSync:
				syncChan <- &inst
			}
		}
	}()
	ticker := time.NewTicker(time.Second * 5)
	nop := guacd.NewInstruction(guacd.InstructionClientNop)
	mouse := guacd.NewInstruction(guacd.InstructionMouse, "0", "2")
	inst1 := nop
	for {
		if inst1.Opcode == mouse.Opcode {
			inst1 = nop
		}
		select {
		case <-ticker.C:
		case syncInst, ok := <-syncChan:
			if !ok {
				return
			}
			//log.Printf("sync instruction: %s\n", syncInst.String())
			inst1 = *syncInst
		}
		//log.Printf("write instruction: %s\n", inst1.String())
		if err2 := tunnel.WriteInstructionAndFlush(inst1); err2 != nil {
			log.Printf("write instruction failed: %s\n", err2)
			return
		}
		//log.Printf("sync tunnel uuid: %s\n", tunnel.UUID)
	}

}

func waitSignal(tunnelCons *sync.Map) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	tick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-c:
			log.Println("receive signal, close all connections")
			disconnect := guacd.NewInstruction(guacd.InstructionClientDisconnect, "close by signal")
			tunnelCons.Range(func(key, value interface{}) bool {
				tunnel := value.(*guacd.Tunnel)
				_ = tunnel.WriteInstructionAndFlush(disconnect)
				if err := tunnel.Close(); err != nil {
					log.Printf("close tunnel failed: %s\n", err)
				} else {
					log.Printf("close tunnel success: %s\n", tunnel.UUID)
				}
				return true
			})
			return
		case <-tick.C:
			count := 0
			tunnelCons.Range(func(key, value interface{}) bool {
				count++
				return true
			})
			log.Printf("current connections: %d\n", count)
		}

	}
}

func randomAddr(a []interface{}, value string) string {
	if len(a) == 0 {
		return value
	}
	rand.Seed(time.Now().UnixNano())
	val := a[rand.Intn(len(a))]
	return val.(string)
}
