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
  * automatic HPKP (multiple keys transparently handled)
  * automatic HSTS
  * Autofinding the Chain of trust to the root certificate
  * no cargo culting required for cipher suit
  * Automatic redirection to HTTPS if a certificate is provided

### HTTP

  * HTTP Range (not Multipart-Ranges)
  * HTTP Cache Compliant
  * wildcard zone redirection (aka redirect `*.some-zo.ne` to `another-zo.ne` )
  * JSON message in AccessLog
  * AccessLog and ErrorLog are syslog RFC 5424 compliant


## Configuration

For the full options list see [conf/config.toml](conf/config.toml) for the main config and [conf/example.vhost](conf/example.vhost)

### Minimal config

```
Listen	= [ "127.0.0.1", "::1" ]
Proxied	= "http://host.tld/"
IncludeVhosts = "/path/to/vhosts"
```

### Minimal vhost with TLS

```
[[Serve]]
zones		= [ "www.f.q.d.n" ]
  [Serve.TLS]
  keys		= [ "/path/privatekey" ]
  cert		= "/path/cert"

[[Redirect]]
From		= [ "f.q.d.n" ]
To		= "www.f.q.d.n"
```

## License
2-Clause BSD


## Todo

  * write comments
  * clean some ugly stuff
  * better handling of CSP configuration (don't be less secure than default)
  * Coping with self signed certificate
  * Coping with onion and local TLD
  * Coping with alert expiration
  * fs-notify for adding / removing configuration files
  * Support OWASP|Naxsi rules
  *
