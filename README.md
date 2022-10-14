# Netlify DynDNS

A little daemon that updates your preferred Netlify DNS settings when your public IP changes

Usage inside a Docker container is preferred as it's easier to manage, but there are Linux and macOS binaries available with every [release][releases].

## Requirements

- Either a:
	1. macOS or a Linux-based operating system;
	2. Or a Docker installation (Preferably).
- A network connection that isn't routed in any particular way through a proxy or VPN.

## Set up with Docker

### With environment variables

```bash
$ docker image pull ghcr.io/jonahgoldwastaken/netlify-dyndns:latest # Images are available for ARM32/64 and X86/X86_64
$ docker run -d \
  --name netlify-dyndns \

  # Required
  -e "NETLIFY_TOKEN=top_secret" \                # API key you created inside the Netlify Admin Panel
  -e "ND_NETLIFY_DOMAIN_NAME=jonahmeijers.nl" \  # The domain name as displayed inside the Netlify Admin Panel
  -e "ND_RECORD_HOSTNAME=home.jonahmeijers.nl" \ # The domain that'll be entered as the hostname on the DNS record

  # Optional
  -e "ND_IP_API=https://api.ipify.org" \         # Custom public IP-lookup API (defaults to 'ipify', must respond with a text body for it to work)
  -e "ND_LOG_LEVEL=info" \                       # The maximum level to log, can be one of "panic", "fatal", "error", "warning", "info", "debug", "trace"
  -e "ND_SCHEDULE=0 0 * * *" \                   # Schedule to run the DNS update at, defaults to every day at 12AM (0 0 * * *)
  -e "ND_RUN_ONCE=false"                         # Set to true to run the DNS update immediately. Scheduling has no effect when run-once is enabled 
  ghcr.io/jonahgoldwastaken/netlify-dyndns
```

> Sourcing ENV variables from a file (with e.g. `docker compose`) is not possible yet.

### With flags

```bash
$ docker image pull ghcr.io/jonahgoldwastaken/netlify-dyndns:latest
$ docker run -d \
  --name netlify-dyndns \
	-e "NETLIFY_TOKEN=top_secret" \
  ghcr.io/jonahgoldwastaken/netlify-dyndns \

  # Required
  "--domain=jonahmeijers.nl" \        # The domain name as displayed inside the Netlify Admin Panel
  "--hostname=home.jonahmeijers.nl" \ # The domain that'll be entered as the hostname on the DNS record

  # Optional
  "--ip-api=https://api.ipify.org" \  # Custom public IP-lookup API (defaults to 'ipify', must respond with a text body for it to work)
  "--log-level=info" \                # The maximum level to log, can be one of "panic", "fatal", "error", "warning", "info", "debug", "trace"
  "--schedule=0 0 * * *" \            # Schedule to run the DNS update at, defaults to every day at 12AM (0 0 * * *)
  "--run-once=false"                  # Set to true to run the DNS update immediately. Scheduling has no effect when run-once is enabled 
```

## Binary

### Installation

With the instructions below you can run `netlify-dyndns` through this command with environment variables either set or flags passed into the executable. Check out the [releases][releases] tab to find the tar name suitable for your system.

```bash
$ curl -Lo netlify-dyndns.tar.gz "https://github.com/jonahgoldwastaken/netlify-dyndns/releases/latest/download/netlify-dyndns_${NEWEST_VERSION}_{Darwin/Linux}_${ARCH}.tar.gz"
$ sudo tar xf netlify-dyndns.tar.gz -C /usr/local/bin netlify-dyndns
```

### Running

```bash
$ netlify-dyndns -h
Automatically update Netlify DNS with your dynamic IP

Usage:
  netlify-dyndns [flags]

Flags:
  -d, --domain string      The domain name registered at Netlify as shown on their dashboard and through their API
  -h, --help               help for netlify-dyndns
  -H, --hostname string    The hostname to be put in the A record
      --ip-api string      The API used to retrieve public IP Address of connected network, must respond with a text string body (default "https://api.ipify.org")
      --log-level string   Maximum level that will be written to stderr (default "info")
      --run-once           Only run the updater once, immediately exiting after
  -S, --schedule string    Cron schedule the DNS check runs on (default "0 0 * * *")
```

[releases]: https://github.com/jonahgoldwastaken/netlify-dyndns/releases
