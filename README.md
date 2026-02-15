# Radxa Quad SATA Hat Controller

This service controls both CPU and Case fans of the
[Quad SATA Hat](https://wiki.radxa.com/Dual_Quad_SATA_HAT) for Raspberry Pi 4B.

> **NOTE**: Only works with Raspberry Pi 4B.

## AAAAAAAAAAAAAAAAAA!!!!!

**WHY THIS PROJECT??**

- No other project I tried worked with both fans.
- I don't need the button and OLED functionality.
- I just need something that turns on the disks and regulates fans.

## Usage

First, put `dtoverlay=pwm-2chan,pin=12,func=4,pin2=13,func2=4` to `config.txt`. This enables pwm
channel 0 and 1 (on pwmchip0) on pins 12 and 13 respectively.

Then, compile with the following command. Make sure that you set `GOOS` and `GOARCH` environment
variables if you are cross compiling.

```sh
go build .
```

You can then run the binary like any other. A fan curve can be modified by passing an argument on
the command line.

```sh
# Contral fans with the curve: "<temp>=<speed>%,<temp>=<speed>%..."
radxa-quad fans --curve '0=0%,70=0%,100=100%'
# Turn on and hold gpio pins that power on gpio disks.
# Notify systemd when disks are ready.
radxa-quad --sdnotify disks --wait-for /dev/sda,/dev/sdb --timeout 2m
```

If you want to run this on startup, create a systemd service. Something like this...

```ini
[Unit]
Description=Quad SATA Hat

[Service]
Type=simple
ExecStart=/path/to/binary
Restart=on-failure

[Install]
WantedBy=multi-user.target
```
