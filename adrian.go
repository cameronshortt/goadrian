package main

import (
    "fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"strconv"

	//"github.com/gokrazy/gokrazy"
	"github.com/spf13/viper"
)

type Printer struct {
	Config  string `mapstructure:config`
	Printer string `mapstructure:printer`
	Webcam  string `mapstructure:webcam`
}

type Printers map[string]Printer

func config() Printers {
	viper.SetConfigFile("/etc/printers.toml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("config error: %w", err))
	}

	var conf Printers
	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatal(fmt.Errorf("config error: %w", err))
	}

	return conf
}

func container(args ...string) {
	cmd := exec.Command("/usr/local/bin/podman", args...)
	cmd.Env = expandPath(os.Environ())
	cmd.Env = append(cmd.Env, "TMPDIR=/tmp")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(fmt.Errorf("container error: %w", err))
	}
}

func fdmmonster(name string, conf Printer) {
	container("run",
		"-td",
		"--name", name,
		"oci-archive:/run/images/fdm-monster");
}

func octoprint(name string, conf Printer) {
	container("run",
		"-td",
		"--name", append(name, "-octoprint"),
		"oci-archive:/run/images/octoprint");
}

func klipper(name string, conf Printer) {
	cpath := append("/etc/", conf.Config)
	rpath := append("/perm/", append(name, "/run"))
	container("run",
		"-td",
		"-v", "/dev:/dev",
		"-v", append(rpath, ":/opt/printer_data/run"),
		"-v", "/perm/gcode:/opt/printer_data/gcodes",
		"-v", append(cpath, ":/opt/printer_data/config/printer.cfg"),
		"--name", append(name, "-klipper"),
		"oci-archive:/run/images/klipper");
}

func mainsail(name string, conf Printers) {
	
}

func moonraker(name string, port int, conf Printer) {
	rpath := append("/perm/", append(name, "/run"))
	container("run",
		"-td",
		"-v", "/dev:/dev",
		"-v", append(rpath, ":/opt/printer_data/run"),
		"-v", "/perm/gcode:/opt/printer_data/gcodes",
		"-v", "/etc/moonraker.conf:/opt/printer_data/config/moonraker.conf"),
		"-p", append(strconv.Itoa(port), ":7125")
		"--name", append(name, "-mainsail"),
		"mkuf/moonraker:latest");
}

func main() {
	conf := config()

	inc := 0
	var port int
	var ports map[string]int
	
	for name, printer := range conf {
	    port = 7125 + inc
	    ports[p] = port

	    go klipper(name, printer)
	    go moonraker(name, port, printer)
	
	    inc += 1
    }

	go mainsail()
}
