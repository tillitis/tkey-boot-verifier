# Distribution Certificate

The distribution certificate is used to sign applications that should
be booted or installed with the boot-verifier. It is a client-side
file format, intended to make it easier to distribute an app and its
corresponding certificate and have compatibility over multiple utility
tools.

See the Go package `distfile` in the [tkey-sign
tool](https://github.com/tillitis/tkey-sign-cli) for more details.

This file format is supported by `tkey-sign tool`, `sign-tool`,
`tkey-mgt` and
[tkeyimage](https://github.com/tillitis/tillitis-key1/tree/main/hw/application_fpga/tools/tkeyimage).

## File format
The distribution certificate consists of a distribution message, named
`message`, a signature and a public key. It is a JSON encoded
structure defined as:

```json
{
  "version": 1,
  "message": {
    "name": <name>,
    "appDigest": [<32 bytes BLAKE2s digest over the application binary>],
    "installable": <boolean>
  },
  "signature": {
    "alg": 1,
    "signature": [<64 bytes signature>]
  },
  "pubkey": {
    "alg": 0,
    "key": [<32 bytes public key>]
  }
}
```

The distribution message is used to create the message to be signed.
This distribution message can be a stand-alone file and used as an
input to `tkey-sign` in order to sign it and create a distribution
certificate.

`sign-tool` can create the distribution message.
