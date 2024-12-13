

## Volumes

Mount volumes to /config

```bash
docker run -d -v $(pwd)/config:/config -p 1812:1812/udp -p 8443:8443/tcp vasilevkirill/twintrustrdp:latest
```

Example Config

```yaml
telegram:
  debug: true
  token: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX:XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
  PoolAddress: 0.0.0.0
  Poolport: 8443
  HookDomain: XXX.example.com
  HookPort: 8443
  admins:
    - 5555555
radius:
  address: 0.0.0.0
  port: 1812
  secret: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
  answertimeout: 30
ldap:
  servers:
    - 192.168.0.100:389
    - 192.168.0.101:389
  user: LDAPUSERNAME
  password: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
  domain: domain.local
  dn: dc=domain,dc=local
cache:
  timeout: 3600
```