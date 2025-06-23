package main

import (
	//"fmt"
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var u_config UserConfig

// Holds information related to general configuration.
type UserConfig struct {
	username    string
	password    string
	regtest     bool
	testnet     bool
	testnet4    bool
	listen_port string
	logfile     string

	public_listen_port string

	prune_ram  uint64
	prune_disk uint64

	unprune_disk bool

	db_dir string

	shutdown_button bool

	chart_proxy string
	init_proxy  string

	deciders_fork bool
}

func load_config() {

	// Setup defaults
	u_config.regtest = false
	u_config.listen_port = "2121"
	u_config.public_listen_port = "21212"
	u_config.chart_proxy = "http://127.0.0.1:12121/"

	// Check for config
	if _, err := os.Stat("config.txt"); err != nil {
		if os.IsNotExist(err) {
			return
		}
	}

	log.Println("Retrieving config...")

	f, conferr := os.Open("config.txt")
	if conferr != nil {
		log.Fatal("config open error: ", conferr)
	}

	rd := bufio.NewReader(f)

	finished := false

	for {
		line, rerr := rd.ReadString('\n')
		if rerr != nil {
			if rerr == io.EOF {
				finished = true
			} else {
				log.Fatal("config read error: ", rerr)
			}
		}

		var tag string
		var info string

		// Read the line until "="
		i := -1
		for ii := range line {
			if string(line[ii]) == "=" {
				i = ii
				break
			}
		}

		switch i <= 0 {
		case true:

		default:
			tag = string(line[:i+1])
			info = strings.TrimRight(string(line[i+1:]), "\r\n")
			log.Println(tag, info)
		}

		switch tag {
		case "indexPrefix=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(uint64(num)) == num {
				index_online_min_prefix = uint64(num)
			}
		case "indexSeed=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(uint32(num)) == num {
				commitment_height_table.Seed = uint32(num)
			}
		case "indexBytes=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(byte(num)) == num {
				commitment_height_table.Bytes = byte(num)
			}
		case "indexMega=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(uint16(num)) == num {
				commitment_height_table.Mega = uint16(num)
			}
		case "indexMaxLoad=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(byte(num)) == num {
				commitment_height_table.MaxLoad = byte(num)
			}
		case "indexHops=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(uint16(num)) == num {
				commitment_height_table.Hops = uint16(num)
			}
		case "indexTweak=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(uint16(num)) == num {
				commitment_height_table.Tweak = uint16(num)
			}
		case "indexMin=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(byte(num)) == num {
				commitment_index_config.Min = byte(num)
			}
		case "indexMax=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(byte(num)) == num {
				commitment_index_config.Max = byte(num)
			}
		case "indexFactor=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(byte(num)) == num {
				index_scaling_factor = byte(num)
			}
		case "btcuser=":
			u_config.username = info
		case "btcpass=":
			u_config.password = info
		case "btcmode=":
			switch info {
			case "regtest":
				u_config.regtest = true
				u_config.testnet = true
			case "testnet":
				u_config.testnet = true
			case "testnet4":
				u_config.testnet = true
				u_config.testnet4 = true
			}
			initTestnet() // init testnet hash if testnet chosen
		case "port=":
			u_config.listen_port = info
		case "public_port=":
			u_config.public_listen_port = info
			if info == "" { // if not running public protocol, set this to 0 to possibly perform better
				index_online_min_prefix = 0
			}
		case "logfile=":
			u_config.logfile = info
		case "pruneRam=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(uint64(num)) == num {
				u_config.prune_ram = uint64(num)
			}
		case "pruneDisk=":
			num, err := strconv.Atoi(info)
			if err != nil {
				log.Println(tag, err)
			} else if int(uint64(num)) == num {
				u_config.prune_disk = uint64(num)
			}
		case "unPruneDisk=", "unpruneDisk=":
			switch info {
			case "true", "TRUE", "True", "Y", "y", "Yes", "yes", "Enabled", "ENABLED", "enabled":
				u_config.unprune_disk = true
			default:
				u_config.unprune_disk = false
			}
		case "db=":
			switch info {
			case "pebble":
				commits_db_backend = CommitsDbBackendPebble
			}
		case "dbdir=":
			u_config.db_dir = info
		case "initproxy=":
			u_config.init_proxy = info
		case "chartproxy=":
			u_config.chart_proxy = info
		case "shutdownButton=", "shutdown_button=":
			switch info {
			case "true", "TRUE", "True", "Y", "y", "Yes", "yes", "Enabled", "ENABLED", "enabled":
				u_config.shutdown_button = true
			default:
				u_config.shutdown_button = false
			}
		case "decidersFork=", "deciders_fork=":
			switch info {
			case "true", "TRUE", "True", "Y", "y", "Yes", "yes", "Enabled", "ENABLED", "enabled":
				u_config.deciders_fork = true
			default:
				u_config.deciders_fork = false
			}
		default:

		}

		if finished {
			break
		}
	}
	f.Close()
}
