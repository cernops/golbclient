## Welcome to the `Lbclient`!

### Application arguments
```bash
Usage:
  lbclient [OPTIONS]

Application Options:
  -l, --log=                   Location of the of log file (default: log/app.log)
  -d, --loglevel=              Console debug level [TRACE, DEBUG, INFO, WARN, ERROR, FATAL] (default: FATAL)
      --fdevel=                File debug level [TRACE, DEBUG, INFO, WARN, ERROR, FATAL] (default: TRACE)
      --flog                   Activate the file logging service
      --cm=                    Set the directory where the client should fetch the configuration files from (default: /usr/local/etc/)
      --ca=                    Set an alternative path for the lbaliases configuration file (default: /usr/local/etc/lbaliases)
  -c, --conf-name=             Set the default name to be used to lookup for the generic configuration file (default: lbclient.conf)
  -v, --version                Version of the file

rotatecfg:
      --rotatecfg.enabled      Enable the automatic rotation of the log files. (default: false)
      --rotatecfg.linelimit=   The maximum amount of lines per log file. (default: 5000)
      --rotatecfg.backuplimit= The maximum amount of backups. (default: 9)
      --rotatecfg.msgBuffer=   The message buffer capacity. (default: 100)

Help Options:
  -h, --help                   Show this help message
```
