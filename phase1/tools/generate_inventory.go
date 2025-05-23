package main

import (
	"gopkg.in/yaml.v3"
	"os"
	"log"
	"fmt"
)

type HostVars struct {
	AnsibleHost string `yaml:"ansible_host"`
	AnsibleUser string `yaml:"ansible_user"`
	Tier				string `yaml:"ansible_tier"`
}

type Hosts map[string]HostVars

type Group struct {
    Hosts Hosts `yaml:"hosts"`
}

type Children struct {
	ControlPlane Group `yaml:"control_plane"`
	Workers Group `yaml:"control_plane"`
}

type Inventory struct {
	All struct {
		Children Children `yaml:"children"`
	} `yaml:"all"`
}

func main(){
	inventory := Inventory{}
	inventory.All.Children.ControlPlane.Hosts = Hosts{
		"k3-master-01"{
			AnsibleHost: "192.168.1.100",
			AnsibleUser: "admin",
			Tier: ""
		},

	}

	inventory.All.Children.Workers.Hosts = Hosts{
		"k3-master-01"{
			AnsibleHost: "192.168.1.100",
			AnsibleUser: "admin",
		},
		"k3-master-01"{
			AnsibleHost: "192.168.1.100",
			AnsibleUser: "admin",
		},
		"k3-master-01"{
			AnsibleHost: "192.168.1.100",
			AnsibleUser: "admin",
		},
	}

	fmt.Println("Started!")
}
