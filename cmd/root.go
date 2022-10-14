// Package cmd contains the root command runnable inside your CLI
package cmd

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonahgoldwastaken/netlify-dyndns/internal/flags"
	"github.com/jonahgoldwastaken/netlify-dyndns/netlify"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = NewRootCommand()

var (
	token        string
	hostname     string
	domain       string
	ipAPI        string
	cronSchedule string
	runOnce      bool
	nClient      *netlify.API
)

// NewRootCommand creates a new root command to run the main application
func NewRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "netlify-dyndns",
		Short:  "Automatically update Netlify DNS with your dynamic IP",
		Run:    Run,
		PreRun: PreRun,
	}
}

func init() {
	flags.Defaults()
	flags.Register(rootCmd)
}

// Execute executues the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// PreRun runs on the root command pre-run hook, setting up the application to be able to run properly
func PreRun(cmd *cobra.Command, _ []string) {
	f := cmd.PersistentFlags()
	log.SetFormatter(&log.TextFormatter{
		EnvironmentOverrideColors: true,
	})

	rawLogLevel, _ := f.GetString(`log-level`)
	if logLevel, err := log.ParseLevel(rawLogLevel); err != nil {
		log.Fatalf("Invalid log level: %s", err.Error())
	} else {
		log.SetLevel(logLevel)
	}

	if err := flags.TestRequired(cmd); err != nil {
		log.Fatal(err)
	}
}

// Run runs on the root command run hook, containing the main business logic for the application
func Run(cmd *cobra.Command, names []string) {
	runLock := make(chan bool, 1)
	runLock <- true

	flags := cmd.PersistentFlags()
	token = viper.GetString("NETLIFY_TOKEN")
	domain, _ = flags.GetString("domain")
	hostname, _ = flags.GetString("hostname")
	ipAPI, _ = flags.GetString("ip-api")
	cronSchedule, _ = flags.GetString("schedule")
	runOnce, _ = flags.GetBool("run-once")

	nClient = netlify.NewAPI(token)

	if runOnce {
		log.Debugln("Running DNS update once")
		runUpdate()
		os.Exit(0)
		return
	}

	if err := runUpdateOnSchedule(cmd, runLock); err != nil {
		log.Error(err)
	}

	os.Exit(1)
}

func runUpdate() {
	zone, err := nClient.GetDNSZoneForDomain(domain)
	if err != nil {
		log.Error(err)
		return
	}
	log.
		WithFields(log.Fields{
			"domain":    domain,
			"zone_id":   zone.ID,
			"zone_name": zone.Name,
		}).
		Debug("Found zone for domain")
	records, err := nClient.GetDNSRecordsForZone(zone.ID)
	if err != nil {
		log.Error(err)
		return
	}
	record, err := nClient.FindDNSForHostname(records, hostname)
	if err != nil {
		log.Error(err)
		return
	}
	log.
		WithFields(log.Fields{
			"id":       record.ID,
			"hostname": record.Hostname,
			"value":    record.Value,
			"ttl":      record.TTL,
		}).
		Debugf("DNS Record fetched from Netlify")
	ip, err := fetchPublicIP(ipAPI)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("Public IP address: %s\n", ip)
	if record.Value == ip {
		log.Info("Current DNS record has the same value as your public IP, aborting update..")
		return
	}

	newRecord := netlify.DNSRecordInput{
		Hostname:   hostname,
		RecordType: "A",
		TTL:        3600,
		Value:      ip,
	}

	if (record == netlify.DNSRecord{}) {
		log.
			WithFields(log.Fields{
				"hostname": newRecord.Hostname,
				"value":    newRecord.Value,
			}).
			Debug("Creating brand new DNS record")
		respRecord, err := nClient.CreateDNSRecord(zone.ID, newRecord)
		if err != nil {
			log.Error(err)
			return
		}

		log.
			WithFields(log.Fields{
				"record": respRecord.ID,
			}).
			Info("Succesfully added new record")
		return
	}

	err = nClient.DeleteDNSRecord(record.DNSZoneID, record.ID)
	if err != nil {
		log.Error(err)
		return
	}

	respRecord, err := nClient.CreateDNSRecord(record.DNSZoneID, newRecord)

	log.WithFields(log.Fields{
		"record": respRecord,
	}).Info("Succesfully added new record")
}

func runUpdateOnSchedule(cmd *cobra.Command, lock chan bool) error {
	log.Debugln("Initialising update schedule")
	if lock == nil {
		lock = make(chan bool, 1)
		lock <- true
	}

	scheduler := cron.New()
	_, err := scheduler.AddFunc(cronSchedule, func() {
		select {
		case v := <-lock:
			defer func() { lock <- v }()
			log.Debugln("Running update")
			runUpdate()
		default:
			log.Debug("Update skipped as another was already in progress")
		}
	})

	if err != nil {
		return err
	}

	scheduler.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(interrupt, syscall.SIGTERM)

	<-interrupt
	scheduler.Stop()
	log.Info("Waiting on update lock..")
	<-lock
	return nil
}

func fetchPublicIP(api string) (string, error) {
	res, err := http.Get(api)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil

}

func logAndExit(err error) {
	log.Error(err)
	os.Exit(1)
}
