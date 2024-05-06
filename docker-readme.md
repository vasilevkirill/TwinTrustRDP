

## Volumes

Mount volumes to /config

```bash
docker run -d -v $(pwd)/config:/config -p 1812:1812/udp -p 8443:8443/tcp vasilevkirill/twintrustrdp:latest
```
