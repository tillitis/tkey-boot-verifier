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

It currently supports:

- Running from app slot 0 on the flash filesystem, verifying the app
  in slot 1. This is the default behaviour.

- Verifying an app from the client. Your client app will typically
  first ask the currently running app to reset and then load the boot
  verifier. Then reset again and load another app. See the `boot`
  command in `tkey-mgt`.

- Installing a device app in slot 1. See the `install` command. This
  needs an installed boot verifier to talk to. The boot verifier *must*
  be installed in slot 0 and its digest noted in firmware, since it
  needs privileged access to the filesystem to be able to install
  apps. See Produce flash image below.

  Right now it automatically resets to start the boot verifier again
  when installation has finished, then it verifies and starts the app
  in slot 1.

- tkey-mgt always sends a reset request to the currently running app
  for both the `boot` and `install` commands. The reset request is
  currently unknown to most apps.

## Build

To build both client app, `tkey-mgt`, and the device app,
`verifier`, run:

```
git submodules init
make
```

To override default behavior and boot into command mode the verifier
app can be built with `BOOT_INTO_WAIT_FOR_COMMAND` defined like so:

```
make EXTRA_CFLAGS=-DBOOT_INTO_WAIT_FOR_COMMAND
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

To install the boot verifier on the flash we use the `tkeyimage` tool in
[tillitis-key1](https://github.com/tillitis/tillitis-key1), but
typically indirectly with make targets. If you just want to create the
flash image file, use:

```
$ cp verifier/app.bin ../tillitis-key1/hw/application_fpga/verifier.bin
$ cp testapp/app_a.bin ../tillitis-key1/hw/application_fpga/
$ cp testapp/app_a.bin.sig ../tillitis-key1/hw/application_fpga/
$ cd ../tillitis-key1/hw/application_fpga/
$ make FLASH_APP_0=verifier.bin FLASH_APP_1=app_a.bin FLASH_APP_1_SIG=app_a.bin.sig flash_image.bin
```

Now you can use `flash_image.bin` with QEMU.

If you want to flash real hardware with the TP-1 programming board,
replace the last command with:

```
$ make FLASH_APP_0=verifier.bin FLASH_APP_1=app_a.bin FLASH_APP_1_SIG=app_a.bin.sig prog_flash
```

You will now have a verifier in app slot 0 and testapp/app_a.bin in
slot 1.

You can try talking to it with `tkey-mgt -cmd install`, see below.

### tkey-mgt

- `tkey-mgt -cmd boot -app path -sig path-to-signature`
- `tkey-mgt -cmd install -app path -sig path-to-signature`

Command `boot` does a verified boot of the device app specified with
`-app`. It assumes a TKey running an app that supports the reset
command.

Command `install` installs the device app specified with `-app` in
slot 1. It assumes you are running an app that supports the reset
command and that a verifier is present in slot 0. See above about
producing a flash image for this use case. It will automatically reset
the TKey after installing, telling firmware to start the verifier on
flash, which will then verify slot 1's digest and reset again to ask
firmware to start slot 1.

You need a signature of the BLAKE2s digest of the app you want to boot
or install. Create this with:

```
$ ./sign-tool -m app -s path-to-private-key
```

The make target `dev-seed` creates a private key seed in `dev-seed`
corresponding to this public key you can use for testing:

```
9b62773323ef41a11834824194e55164d325eb9cdcc10ddda7d10ade4fbd8f6d
```

NOTE WELL: This will most likely move to the [tkey-sign
tool](https://github.com/tillitis/tkey-sign-cli).

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
