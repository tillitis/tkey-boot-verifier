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
  in slot 1 using the pubkey also stored on flash. This is the default
  behaviour.

- Installing a device app in slot 1. See the `install` command. This
  needs an installed boot verifier to talk to. The boot verifier *must*
  be installed in slot 0 and its digest noted in firmware, since it
  needs privileged access to the filesystem to be able to install
  apps. See Produce flash image below.

  Right now it automatically resets to start the boot verifier again
  when installation has finished, then it verifies and starts the app
  in slot 1.

- Installing a vendor pubkey on flash.

- Running both verifier and a verified device app sent from client.

- To start tkey-boot-verifier, tkey-mgt always sends a reset request
  to the currently running app. The reset request is currently unknown
  to most apps.

## Build

To build both client app, `tkey-mgt`, and the device app,
`verifier`, run:

```
git submodule update --init --recursive
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

### Produce flash image or flash real hardware

To install the boot verifier on the flash we use the `tkeyimage` tool in
[tillitis-key1](https://github.com/tillitis/tillitis-key1), but
typically indirectly with make targets. If you just want to create the
flash image file, use:

```
$ cp verifier/app.bin ../tillitis-key1/hw/application_fpga/verifier.bin
$ cp testapp/app_a.bin ../tillitis-key1/hw/application_fpga/
$ cp testapp/app_a.bin.sig ../tillitis-key1/hw/application_fpga/
$ cp testapp/pubkey ../tillitis-key1/hw/application_fpga/
$ cd ../tillitis-key1/hw/application_fpga/
$ make FLASH_APP_0=verifier.bin FLASH_APP_1=app_a.bin FLASH_APP_1_SIG=app_a.bin.sig FLASH_APP_1_PUB=pubkey flash_image.bin
```

You will now have the boot verifier in app slot 0 and
`testapp/app_a.bin` in slot 1.

If you want to flash real hardware with the TP-1 programming board,
replace the last command with:

```
$ make FLASH_APP_0=verifier.bin FLASH_APP_1=app_a.bin FLASH_APP_1_SIG=app_a.bin.sig FLASH_APP_1_PUB=pubkey prog_flash
```

You can try talking to the device (emulated or not) with `tkey-mgt
-cmd install`, see below.

### Using QEMU

You can use the `flash_image.bin` file with QEMU, but you have to
build the QEMU-specific firmware first. Point out you want the boot
verifier in slot 0 and build like this:

```
make FLASH_APP_0=verifier.bin qemu_firmware.elf
```

Remember that you need to use the `qemu/tools/tk1/qemu_usb_mux.py`
script from our [QEMU repo](https://github.com/tillitis/qemu) to be
able to talk to the firmware/apps when using QEMU.

### tkey-mgt

- `tkey-mgt [-no-expect-close] -cmd boot -app path -sig path-to-signature -pub path-to-pubkey`
- `tkey-mgt [-no-expect-close] -cmd install -app path -sig path-to-signature`
- `tkey-mgt [-no-expect-close] -cmd install-pubkey -pub path`

*NB*: use `-no-expect-close` when running `tkey-mgt` against QEMU. The
connection behaves differently compared to real hardware.

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

Command `install-pubkey` installs the pubkey specified with `-pub`,
replacing any installed pubkey. During the installion the user is
asked to confirm by touching the TKey touch sensor three times.

A pubkey file can be created with:

```
$ ./sign-tool -p pubkey -s path-to-private-key
```

NOTE WELL: For real use signing of device apps [the tkey-sign
tool](https://github.com/tillitis/tkey-sign-cli) with BLAKE2s support
will most likely be used instead of `sign-tool`.

## Chained Reset

### Example: Verified boot from client

Given:

- A verifier app X in preloaded app slot 0
- An app A in preloaded app slot 1
- A verifier app Y on the client
- An app B on the client

The following reset chain can be used to, from the client, first load a
verifier and then load a verified app.

| *reset type*              | *app digest*  | *next app data*         | *next app*                 |
|---------------------------|---------------|-------------------------|----------------------------|
| START_DEFAULT (Cold boot) | H(verifier X) | 000...                  | Verifier X from slot 0     |
| START_FLASH1_VER          | H(app A)      | -                       | App A from slot 1          |
| START_CLIENT_VER          | H(verifier Y) | BV_NAD_WAIT_FOR_COMMAND | Verifier Y from client     |
| START_CLIENT_VER          | H(app B)      | -                       | App B from client          |

## Verifier application protocol

`verifier` has a simple application protocol on top of the [TKey
Framing Protocol](https://dev.tillitis.se/protocol/#framing-protocol).

The protocol has the following requests and responses:

| *command*              | *function*                                         | *length* | *code* | *data*                                              | *response*             |
|------------------------|----------------------------------------------------|----------|--------|-----------------------------------------------------|------------------------|
| `CMD_VERIFY`           | Verify an app signature and reset into client mode | 128 B    | 0x01   | 32 B next apps digest, 64 B signature               | none                   |
| `CMD_UPDATE_APP_INIT`  | Initialize app installation                        | 128 B    | 0x03   | 32 bit LE app size, 32 B app digest, 64 B signature | `CMD_UPDATE_APP_INIT`  |
| `CMD_UPDATE_APP_CHUNK` | Store a chunk of an app on flash                   | 128 B    | 0x04   | 127 B app data                                      | `CMD_UPDATE_APP_CHUNK` |
| `CMD_GET_PUBKEY`       | Get the public key installed on flash              | 1 B      | 0x05   | none                                                | `CMD_GET_PUBKEY`       |
| `CMD_STORE_PUBKEY`     | Store public key on flash                          | 128 B    | 0x06   | 32 B public key                                     | `CMD_STORE_PUBKEY`     |
| `CMD_SET_PUBKEY`       | Set pubkey used by `CMD_VERIFY`                    | 128 B    | 0x07   | 32 B public key                                     | `CMD_SET_PUBKEY`       |
| `CMD_ERASE_AREAS`      | Erase all app storage areas                        | 1 B      | 0x08   | none                                                | `CMD_ERASE_AREAS`      |
| `CMD_RESET`            | Reset TKey                                         | 4 B      | 0xfe   | 1 B reset type, 1 B next app data                   | none                   |

| *response*             | *length* | *code* | *data*                       |
|------------------------|----------|--------|------------------------------|
| `CMD_UPDATE_APP_INIT`  | 4 B      | 0x03   | 1 B status                   |
| `CMD_UPDATE_APP_CHUNK` | 4 B      | 0x04   | 1 B status                   |
| `CMD_GET_PUBKEY`       | 128 B    | 0x05   | 1 B status + 32 B public key |
| `CMD_STORE_PUBKEY`     | 4 B      | 0x06   | 1 B status                   |
| `CMD_SET_PUBKEY`       | 4 B      | 0x07   | 1 B status                   |
| `CMD_ERASE_AREAS`      | 4 B      | 0x08   | 1 B status                   |

| *status replies* | *code* |
|------------------|--------|
| OK               | 0      |
| BAD              | 1      |

Digests are computed using BLAKE2s with 32-byte digest size.
Signatures and public keys use ed25519.

Please note that `verifier` also replies with a `NOK` Framing Protocol
response status if the endpoint field in the FP header is meant for
the firmware (endpoint = `DST_FW`). This is recommended for
well-behaved device applications so the client side can probe for the
firmware.

### Commands

#### `CMD_ERASE_AREAS`

Erases and deallocates all app data storage areas. Requires user
presence confirmation (touch) three times.

Response:

- `STATUS_OK`
- `STATUS_BAD`: User presence confirmation failed or erase operation
  failed.

#### `CMD_GET_PUBKEY`

Retrieves the stored public key.

Response:

- `STATUS_OK`
- `STATUS_BAD`: Failed to retrieve public key.

#### `CMD_STORE_PUBKEY`

Stores a public key onto the TKey. Any previously installed pubkey is
replaced. Requires user presence confirmation (touch) three times.

Response:

- `STATUS_OK`
- `STATUS_BAD`: User presence failed or storage operation failed.

#### `CMD_SET_PUBKEY`

Sets the temporary vendor public key used by the `CMD_VERIFY` command.

Response:

- `STATUS_OK`

#### `CMD_VERIFY`

Verifies the provided signature against the provided digest and the
public key previously set using the `CMD_SET_PUBKEY` command.

No response is sent. If verification succeeds, the device is reset
into verified-app-from-client mode, with the allowed app digest set to
the verified digest. If verification fails, the device is halted.

#### `CMD_RESET`

Performs a system reset using the provided reset type and next app
data.

Calls the `sys_reset(struct reset *rst)` syscall with the provided
data in `rst` like so:

```
rst->type = reset type
rst->next_app_data = next app data
```

No response. The device resets immediately.

#### `CMD_UPDATE_APP_INIT`

Initializes application installation. Requires user presence
confirmation (touch) three times.

After this command succeeds the currently installed app has been
erased and the client is only allowed to send `CMD_UPDATE_APP_CHUNK`
commands. The internal update address counter used by
`CMD_UPDATE_APP_CHUNK` will be set to 0.

Response:

- `STATUS_OK`
- `STATUS_BAD`: User presence failed or initialization failed.

#### `CMD_UPDATE_APP_CHUNK`

Stores a chunk of an application onto the TKey. The chunk will be
written to the address pointed to by the internal update address
counter.

After this command succeeds the size of the provided data has been
added to the internal update address counter. When the last chunk has
been stored, the device will automatically reset and the installed app
started.

All commands must contain 127 bytes of data. If the size of the last
chunk is not 127, zero-pad until it is.

Response:

- `STATUS_OK`
- `STATUS_BAD`: Storing app chunk failed.

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
