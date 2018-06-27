package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/studio-b12/gowebdav"
)

const configPath = "downloader.toml"

type Show struct {
	Name    string
	Seasons []int
}

type Config struct {
	Source      string
	Destination string
	Shows       []Show `toml:"Show"`
}

func escape(path string) string {
	return path
	//return url.PathEscape(path)
	//strings.Replace(path, " ", "+", -1)
}

func main() {
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		log.Fatalf("Failed to decode config file %s: %v\n", configPath, err)
	}

	if _, err := os.Stat(config.Destination); err != nil {
		log.Fatalf("Destination directory <%s> cannot be found: %v\n", config.Destination, err)
	}

	client := gowebdav.NewClient(config.Source, "", "")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	for _, show := range config.Shows {
		for _, season := range show.Seasons {
			log.Printf("Looking at show <%s> season <%d>\n", show.Name, season)

			srcRoot := fmt.Sprintf("/%s/S%.2d/", show.Name, season)
			dstRoot := fmt.Sprintf("%s/%s/S%.2d/", config.Destination, show.Name, season)

			if err := os.MkdirAll(dstRoot, os.ModePerm); err != nil {
				log.Fatalf("Failed to create directory <%s>: %v\n", dstRoot, err)
			}

			log.Printf("Comparing src <%s> to dst <%s>\n", srcRoot, dstRoot)

			files, err := client.ReadDir(srcRoot)
			if err != nil {
				log.Printf("Failed to read src: %v", err)
				continue
			}

			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".mp4") {
					log.Println(file.Name())

					srcFile := srcRoot + "/" + escape(file.Name())
					dstFile := dstRoot + "/" + file.Name()

					if _, err := os.Stat(dstFile); os.IsNotExist(err) {
						log.Printf("Downloading <%s>\n", srcFile)
						data, err := client.Read(srcFile)
						if err != nil {
							log.Fatalf("Failed to read file <%s>: %v\n", srcFile, err)
						}

						log.Printf("Writing <%s>\n", dstFile)
						if err := ioutil.WriteFile(dstFile, data, os.ModePerm); err != nil {
							log.Fatalf("Failed to write file <%s>: %v\n", dstFile, err)
						}
					}
				}
			}
		}
	}
}
