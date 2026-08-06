# poggadaj-HTTP
A minor part of the server stack, currently the only thing implemented is returning the static IP of the TCP server to Gadu-Gadu 6.1 clients and adserver.

Currently, it has to be run on port 80, so you need root permissions.

To connect to it, add the following to the hosts file:
```
127.0.0.1 appmsg.gadu-gadu.pl
127.0.0.1 adserver.gadu-gadu.pl
```
