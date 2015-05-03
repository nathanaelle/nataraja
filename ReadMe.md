= Nataraja


== What is Nataraja ?

Nataraja is a HTTP reverse proxy.


== Why another one ?

the main goals are :

  * simple configuration
  * Default secure TLS without any knowledge
  * rfc5424 syslog support
  * Minimal WAF (TODO)
  * transparent reload conf (TODO)


== Configuration

see [conf/config.toml] for the main config and [conf/example.vhost]

== License
2-Clause BSD

== Todo

  * write comments
  * clean some ugly stuff
  * better handling of HPKP, HSTS, CSP (don't be less secure than default)
