package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/nikola43/stardust/config"
	crypto "github.com/nikola43/stardust/crypto"
	"github.com/nikola43/stardust/router"
	"github.com/nikola43/stardust/server"
	sysinfo "github.com/nikola43/stardust/sysinfo"
)

var (
	configFile string
	update     string
	etcd       bool
	cfg        *config.Config
)

func init() {
	flag.StringVar(&configFile, "config", "", "configuration file")
	flag.StringVar(&update, "update", "", "update etc / file")
	flag.BoolVar(&etcd, "etcd", false, "enable etcd")
	flag.Parse()
}

func main() {

	sysHash := GetSysInfo()

	key := make([]byte, 32)
	rand.Read(key)
	fmt.Println(key)
	crypto.EncryptSysData([]byte(sysHash), []byte(key))
	crypto.DecryptFile("sysdata.txt.bin", []byte(key))
	os.Exit(0)

	// create unix syscall
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	notify := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	UpdateEtcdConf()

	// create unix syscall
	signal.Notify(sig, os.Interrupt, os.Kill)

	// get etcd config
	cfg := GetEtcdConfig()
	cfg.Watcher(ctx, notify)

	// init server
	r := router.New(ctx)
	s := server.Server{
		Config: cfg,
		Router: r,
		Notify: notify,
	}
	s.Run(ctx)
	<-sig
}

func InitServer(octx context.Context, notify *chan struct{}) {
	r := router.New(octx)
	s := server.Server{
		Config: cfg,
		Router: r,
		Notify: *notify,
	}

	s.Run(octx)
}

func UpdateEtcdConf() {
	// check if we need update nodes config file
	if update != "" {
		err := config.UpdateConf(update, configFile)
		if err != nil {
			fmt.Println("UpdateConf")
			panic(err)
		}
		os.Exit(0)
	}
}

func GetEtcdConfig() *config.Config {
	var cfg *config.Config

	// get etcd config
	if etcd {
		cfg = config.New().FromEtcd(configFile)
	} else {
		cfg = config.New().FromFile(configFile)
	}
	err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func GetSysInfo() string {

	info := sysinfo.NewSysInfo()
	fmt.Printf("%+v\n", info)
	fmt.Printf("%+s\n", info.ToString())
	fmt.Printf("%+s\n", info.ToHash())
	return info.ToHash()
}
