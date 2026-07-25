# Name: TBD

My project to control my home infrastructure.

Right now it consist of thinkpad on Ubuntu server and Raspberry pi as main home gateway.

## Features

- Turn on and off thinkpad
  - if user in local network, use local magic packet
  - if user not in local network, use ssh pi and send magic packet from there
- Check if thinkpad is up