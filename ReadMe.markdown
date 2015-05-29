# Nataraja


## What is Nataraja ?

Nataraja is a HTTP reverse proxy.


## Why another one ?

### Main goals

  * simple configuration
  * Default secure TLS without any knowledge
  * Minimal WAF (TODO)
  * transparent reload conf (TODO)
  * HTTP Caching (TODO)

### Rejected goals

  * coping with <any>cgi
  * permit TLS tuning
  * Non "reverse proxy" proxy
  * Serving static files
  * ESI scripting

## Features

### HTTPS

  * automatic OCSP
  * automatic HPKP
  * automatic HSTS
  * Autofinding the Chain of trust to the root certificate
  * no cargo culting required for cipher suit
  * Automatic redirection to HTTPS if a certificate is provided


## Configuration

see [conf/config.toml] for the main config and [conf/example.vhost]


## License
2-Clause BSD


## Todo

  * write comments
  * clean some ugly stuff
  * better handling of HPKP, HSTS, CSP configuration (don't be less secure than default)
  * Coping with self signed certificate
  * Coping with onion and local TLD
  * Coping with alert expiration
  * fs-notify for adding / removing configuration files
  * Support OWASP|Naxsi rules
  *
