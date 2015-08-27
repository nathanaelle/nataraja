# Nataraja


## What is Nataraja ?

Nataraja is a HTTP reverse proxy.


## Why another one ?

### Main goals

  * [x] simple configuration
  * [x] Default secure TLS without any knowledge
  * [x] transparent support of HTTP/1.1 & HTTP/2
  * [ ] Minimal WAF
  * [ ] transparent reload conf
  * [ ] HTTP Caching

### Rejected goals

  * coping with <any>cgi
  * permit TLS tuning
  * Non "reverse proxy" proxy
  * Serving static files
  * ESI scripting


## Configuration

For the full options list see [conf/config.toml](conf/config.toml) for the main config and [conf/example.vhost](conf/example.vhost)

### Minimal config

```
Listen	= [ "127.0.0.1", "::1" ]
Proxied	= "http://host.tld/"
IncludeVhosts = "/path/to/vhosts"
```

### Minimal vhost without TLS

```
[[Serve]]
zones		= [ "www.f.q.d.n" ]

[[Redirect]]
From		= [ "f.q.d.n" ]
To		= "www.f.q.d.n"
```

### Minimal vhost with TLS

```
[[Serve]]
zones		= [ "www.f.q.d.n" ]
[Serve.TLS]
keys		= [ "/path/privatekey-1", "/path/publickey-2" ]  # the second key is a public key only for HPKP
cert		= "/path/cert"

[[Redirect]]
From		= [ "f.q.d.n" ]
To		= "www.f.q.d.n"
```


## Features

### HTTPS

  * automatic OCSP
  * automatic HPKP (multiple public and private keys transparently handled)
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


## License
2-Clause BSD


## Todo

  * write comments
  * better handling of CSP configuration (don't be less secure than default)
  * Coping with self signed certificate
  * Coping with onion and local TLD
  * Coping with alert expiration
  * fs-notify for adding / removing configuration files
  * Support OWASP|Naxsi rules
