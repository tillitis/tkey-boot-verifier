# tkey-boot-verifier

**WARNING**: Work in progress!

The TKey boot verifier is a boot stage for the Tillitis TKey. With the
support of the TKey firmware it implements a combination of measured
boot and verified boot which makes it possible to upgrade the verified
app without losing data, including the cryptographic keys.

The boot verifier can start either from slot 0 in flash or be loaded
by a client app.

See [the design document](doc/design.md) for more.

## Status

For test purposes the boot verifier currently waits for commands from the
client after power cycling. You can talk to it with `tkey-mgt`.

It currently supports:

- verifying an app from the client. Your client app will typically
  first load the boot verifier, then resetting and loading another
  app. See the `boot` command in `verifier-client`. Note that this
  needs a Castor TKey using the `defaultapp` in slot 0, which makes it
  reset to firmware, waiting for commands from the client.

- installing a device app in slot 1. See the `install` command. This,
  on the other hand, needs a running boot verifier to talk to. The
  boot verifier *must* be installed in slot 0 and its digest noted in
  firmware, since it needs privileged access to the filesystem to be
  able to install apps. See Produce flash image below.

  Right now it automatically resets to start the boot verifier again
  when installation has finished, then it verifies and starts the app
  in slot 1.

## Build

To build both client app, `tkey-mgt`, and the device app,
`verifier`, run:

```
./build.sh
```

## Use

For all uses of the boot verifier, you need to build [a current Castor
TKey](https://github.com/tillitis/tillitis-key1) which by default uses
the `defaultapp` in slot 0. Flash it on a TKey Unlocked with the TKey
Programmer Board. Buy here:

https://shop.tillitis.se/

You will also need [a CH55x reset
controller](https://shop.blinkinlabs.com/products/ch55x-reset-controller)
to update the USB controller firmware to Castor.

See [TKey Developer
Handbook](https://dev.tillitis.se/castor/unlocked/) for instructions.

The boot verifier can be placed on flash, typically slot 0, then verifying
slot 1, or loaded by a client app, followed by the app to be verified.

### Produce flash image

To install the boot verifier on the flash, use the `tkeyimage` tool in
[tillitis-key1](https://github.com/tillitis/tillitis-key1). Typically:

```
$ cp verifier/app.bin ../tillitis-key1/hw/application_fpga/
$ cd ../tillitis-key1/hw/application_fpga/
$ ./tools/tkeyimage/tkeyimage -f -app0 verifier.bin -o flash_image.bin
$ make FLASH_APP_0=verifier.bin prog_flash
```

You will now have a boot verifier in app slot 0. In the current state of
development it will wait for commands from the client after starting.
This is not the end goal, but sufficient for development.

You can try talking to it with `tkey-mgt -cmd install`, see
below.

### tkey-mgt

- `tkey-mgt -cmd boot -app path`
- `tkey-mgt -cmd install -app path`

Command `boot` does a verified boot of the device app specified with
`-app`. It assumes a TKey running firmware which is waiting for
commands from client. In the current state of development this
typically means a Castor prototype with the `defaultapp` in app slot
0.

Command `install` installs the device app specified with `-app` in
slot 1. It assumes you are running a boot verifier from slot 0 which is
waiting for commands from the client. See above about producing a
flash image for this use case. It will automatically reset the TKey
after installing, telling firmware to start the boot verifier on flash,
which will then verify slot 1's digest and reset again to ask firmware
to start slot 1.

## Chained Reset

### Example: Verified boot from client

Given:

- A verifier app X in preloaded app slot 0
- An app A in preloaded app slot 1
- A verifier app Y on the client
- An app B on the client

The following reset chain can be used to, from the client, first load a
verifier and then load a verified app.

| Reset Type                | App Digest    | Next App Data           | Next app                   |
|---------------------------|---------------|-------------------------|----------------------------|
| START_DEFAULT (Cold boot) | H(verifier X) | 000...                  | Verifier X from slot 0     |
| START_FLASH1_VER          | H(app A)      | -                       | App A from slot 1          |
| START_CLIENT_VER          | H(verifier Y) | BV_NAD_WAIT_FOR_COMMAND | Verifier Y from client     |
| START_CLIENT_VER          | H(app B)      | -                       | App B from client          |

## TODO

- Change default behaviour of boot verifier to always start app slot
  1, instead of waiting for commands.
  
- When installing an app in slot 1, always reset digest and signature
  first, and detect it on start, so we can resume a botched
  installation.

- Change state machine to:

  ```mermaid
  stateDiagram-v2
    [*] --> INIT
    INIT --> VERIFY_FLASH: next_app_data == 17
    INIT --> WAIT_FOR_COMMAND: next_app_data != 17
    VERIFY_FLASH --> BOOT
    WAIT_FOR_COMMAND --> WAIT_FOR_APP_CHUNK: CMD_UPDATE_APP_INIT
    WAIT_FOR_APP_CHUNK --> WAIT_FOR_APP_CHUNK: CMD_UPDATE_APP_CHUNK
    WAIT_FOR_APP_CHUNK --> BOOT: Last CMD_UPDATE_APP_CHUNK
    BOOT --> [*]
  ```

## Licenses and SPDX tags

Unless otherwise noted, the project sources are copyright Tillitis AB,
licensed under the terms and conditions of the "BSD-2-Clause" license.
See [LICENSE](LICENSE) for the full license text.

External source code we have imported are isolated in their own
directories. They may be released under other licenses. This is noted
with a similar `COPYING`/`LICENSE` file in every directory containing
imported sources.

The project uses single-line references to Unique License Identifiers
as defined by the Linux Foundation's [SPDX project](https://spdx.org/)
on its own source files, but not necessarily imported files. The line
in each individual source file identifies the license applicable to
that file.

The current set of valid, predefined SPDX identifiers can be found on
the SPDX License List at:

https://spdx.org/licenses/

We attempt to follow the [REUSE
specification](https://reuse.software/).
