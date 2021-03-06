package main

import (
	"github.com/coreos/bbolt"
	lightning "github.com/fiatjaf/lightningd-gjson-rpc"
	"github.com/fiatjaf/lightningd-gjson-rpc/plugin"
)

const (
	DATABASE_FILE = "chanstore.db"
)

var (
	continuehook = map[string]string{"result": "continue"}
	multipliers  = map[string]float64{
		"msat": 1,
		"sat":  1000,
		"btc":  100000000000,
	}
	channelswaitingtosend = map[string]*lightning.Channel{}
	serverlist            = make(map[string]string)
)

var db *bbolt.DB
var err error

func main() {
	p := plugin.Plugin{
		Name:    "chanstore",
		Version: "v0.1",
		Dynamic: true,
		Options: []plugin.Option{
			{
				"chanstore-connect",
				"string",
				"",
				"Chanstore service addresses to fetch channels from, comma-separated.",
			},
			{
				"chanstore-server",
				"bool",
				false,
				"If enabled, run a chanstore server.",
			},
			{
				"chanstore-price",
				"int",
				72,
				"Satoshi price to ask for peers to include a channel.",
			},
		},
		Hooks: []plugin.Hook{
			{
				"rpc_command",
				func(p *plugin.Plugin, payload plugin.Params) (resp interface{}) {
					rpc_command := payload.Get("rpc_command")

					switch rpc_command.Get("method").String() {
					case "getroute":
						return getroute(p, rpc_command)
					case "listchannels":
						return listchannels(p, rpc_command)
					case "dev-sendcustommsg":
						p.Log("dev-sendcustommsg: ", payload.Get("rpc_command").String())

						return map[string]interface{}{
							"result": "continue",
						}
					default:
						return continuehook
					}
				},
			},
			{
				"custommsg",
				custommsg,
			},
		},
		OnInit: onInit,
	}

	p.Run()
}
