// Netlify-dyndns enables dynamic dns for a domain name hosted on Netlify
package main

import (
	"github.com/jonahgoldwastaken/netlify-dyndns/cmd"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func main() {
	cmd.Execute()
}
