# Nataraja


## What is Nataraja ?

Nataraja is a HTTP reverse proxy.


## Why another one ?

### Main goals

  * [x] Simple configuration
  * [x] Default secure TLS without any knowledge
  * [x] transparent support of HTTP/1.1 & HTTP/2
  * [x] IPv4 & IPv6
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
zones           = [ "www.f.q.d.n" ]

[[Redirect]]
From            = [ "f.q.d.n" ]
To              = "www.f.q.d.n"
```

### Minimal vhost with TLS


```
[[Serve]]
zones           = [ "www.f.q.d.n" ]
# Only one key is acceptable because HSTS pinning/preload is disabled by default
keys            = [ "/path/privatekey-1" ]
cert            = "/path/cert"

[[Redirect]]
From            = [ "f.q.d.n" ]
To              = "www.f.q.d.n"
```


### A more Complicated Case

```
[[Serve]]
zones           = [ "www.foo.example", "foo.example" ]
keys            = [ "/etc/ssl/foo.priv-1", "/etc/ssl/foo.pub-2" ]  # the second key is a public key only for HPKP
cert            = "/path/foo.cert"

[[Serve]]
zones                   = [ "www.bar.example" ]
StrictTransportSecurity = "max-age=15552000;includeSubdomains"
XFrameOptions		= "DENY"
keys                    = [ "/etc/ssl/bar.priv-1", "/etc/ssl/bar.pub-2" ]  # the second key is a public key only for HPKP
cert                    = "/path/bar.cert"

[[Serve]]
zones           = [ "quux.example" ]

[[Redirect]]
From            = [ "bar.example" ]
To              = "www.bar.example"

[[Redirect]]
From            = [ ".quux.example", "another-quux.example", ".another-quux.example" ]
To              = "quux.example"

```

### Explanation for the TLS configuration

the TLS configuration for a virtual host, needs 2 mandatory informations :

  * the path to the certificate provided by the CA
  * the path to the private key
  * optionaly, you can add severals public keys to the primary key

all the other configurations fall in two cases :

  * good practices like cipher suite, accepted protocols, HSTS, …
  * informations already stored in the certificate like OCSP responder, intermediate certificate, …

so, it seems logical to provide only the non guessable part : the certificates and the private and public keys


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
