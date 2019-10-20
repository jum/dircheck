dircheck.go
===========

A small command line utility to quickly check for changes to a set of
directories written in the [go](http://golang.org) language.

For example a ~/Library/LaunchAgent file:

```plist
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN"  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>org.mager.launchctl</string>

  <key>LowPriorityIO</key>
  <true/>

  <key>Program</key>
  <string>/Users/jum/bin/launchctl.sh</string>

  <key>WatchPaths</key>
  <array>
    <string>/Users/jum/Library/LaunchAgents</string>
    <string>/Library/LaunchAgents</string>
    <string>/Library/LaunchDaemons</string>
    <string>/System/Library/LaunchAgents</string>
    <string>/System/Library/LaunchDaemons</string>
  </array>

  <key>RunAtLoad</key>
  <true/>
</dict>
</plist>
```

With the associated shell script:

```sh
#!/bin/sh
PATH=$HOME/gopkg/bin:$PATH
(
	echo `date +'%Y%m%dT%H%M%S'`
	dircheck -f ~/.dircheck_launch ~/Library/LaunchAgents /Library/Launch* /System/Library/Launch*
) | open -f
exit 0
```

A console will popup as soon as there are any changes to the launch daemon
config files.
