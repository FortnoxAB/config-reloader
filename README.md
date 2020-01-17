# config-reloader

Used in kubernetes for detecting changes in mounted configmaps
Usage example:
```
#signal USR2 to pid 1 when /tmp/foo changes
config-reloader -watch /tmp/foo:USR2:1

```
