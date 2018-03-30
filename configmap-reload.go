package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	fsnotify "gopkg.in/fsnotify.v1"
)

var volumeDirs volumeDirsFlag
var command []string

func main() {
	flag.Var(&volumeDirs, "volume-dir", "the config map volume directory to watch for updates; may be used multiple times")
	flag.Parse()

	if len(volumeDirs) < 1 {
		log.Println("Missing volume-dir")
		log.Println()
		flag.Usage()
		os.Exit(1)
	}

	command = flag.Args()
	if len(command) == 0 {
		log.Fatalf("A command argument is required.")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					if filepath.Base(event.Name) == "..data" {
						log.Println("config map updated")
						run(command)
						log.Println("successfully triggered reload")
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	for _, d := range volumeDirs {
		log.Printf("Watching directory: %q", d)
		err = watcher.Add(d)
		if err != nil {
			log.Fatal(err)
		}
	}
	<-done
}

type volumeDirsFlag []string

func (v *volumeDirsFlag) Set(value string) error {
	*v = append(*v, value)
	return nil
}

func (v *volumeDirsFlag) String() string {
	return fmt.Sprint(*v)
}

func run(command []string) error {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
