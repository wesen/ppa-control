# PPA Control

This application allows simple preset changing and master volume management
for DSP boards by PPA.

It allows handles discovery by sending out UDP broadcast packets.

## Command line usage

### Starting a simulated speaker

To simulate a speaker, run:

```shell
go run ./cmd/ppa-cli simulate --address 0.0.0.0 --log-level info
```

### Pinging and discovering speaker

To ping and discover speakers locally, run:

```shell 
go run ./cmd/ppa-cli ping --log-level info --discover
```

## UI usage

To run the UI:

```shell
go run ./cmd/ui-test
```

To build the UI, first install [fyne](https://developer.fyne.io/started/),
then run `make`.

```shell 
make
```