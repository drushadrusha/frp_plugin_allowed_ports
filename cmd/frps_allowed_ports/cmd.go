package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gofrp/fp-multiuser/pkg/server"

	"github.com/spf13/cobra"
)

const version = "1.0.2"

var (
	showVersion bool
	bindAddr    string
	// tokenFile   string
	portsFile string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version")
	rootCmd.PersistentFlags().StringVarP(&bindAddr, "bind_addr", "l", "127.0.0.1:9000", "bind address")
	rootCmd.PersistentFlags().StringVarP(&portsFile, "ports_file", "p", "./ports", "ports file")
}

var rootCmd = &cobra.Command{
	Use:   "fp-multiuser",
	Short: "fp-multiuser is the server plugin of frp to support multiple users.",
	RunE: func(cmd *cobra.Command, args []string) error {

		portslist, _ := ParseportsFromFile(portsFile)
		if showVersion {
			fmt.Println(version)
			return nil
		}
		s, err := server.New(server.Config{
			BindAddress: bindAddr,
			Ports:       portslist,
		})
		if err != nil {
			return err
        }
        go s.Run()

        wd, err := os.Getwd()
        if err != nil {
            log.Printf("Error getting current working directory: %v\n", err)
            return err
        }
        absolutePortsFile := filepath.Join(wd, portsFile)

        log.Println("Starting file watcher goroutine...")
        go watchPortsFile(s, absolutePortsFile)

        // Keep the main goroutine alive
        for {
            time.Sleep(time.Second)
        }
    },
}

func watchPortsFile(s *server.Server, portsFile string) {
	log.SetOutput(os.Stdout)
	log.Printf("watchPortsFile received path: %s\n", portsFile)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create file watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	log.Printf("Watching file: %s\n", portsFile)
	err = watcher.Add(portsFile)
	if err != nil {
		log.Printf("Failed to watch ports file: %v\n", err)
		return
	}
	log.Printf("Successfully added watcher for file: %s\n", portsFile)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Printf("Event: %s, Name: %s, Op: %s\n", event.String(), event.Name, event.Op.String())
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Reloading ports file...")
				portslist, err := ParseportsFromFile(portsFile)
				if err != nil {
					log.Printf("Error reloading ports file: %v\n", err)
					continue
				}
				s.Reload(portslist)
			} else if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				log.Printf("Ports file %s removed or renamed. Re-adding watcher.\n", portsFile)
				// If the file is removed or renamed, we need to re-add the watcher
				err = watcher.Add(portsFile)
				if err != nil {
					log.Printf("Failed to re-add watcher for ports file: %v\n", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error watching ports file: %v\n", err)
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func ParseportsFromFile(file string) (map[string][]string, error) {
	buf, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]string)
	rows := strings.Split(string(buf), "\n")
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		parts := strings.Split(row, "=")
		if len(parts) != 2 {
			log.Printf("Skipping invalid line in ports file: %s\n", row)
			continue
		}
		key := parts[0]
		value := parts[1]
		result[key] = append(result[key], value)
	}
	return result, nil

}
