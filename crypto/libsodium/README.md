# libsodium

A modern, portable, easy to use crypto library.

Imported from https://github.com/jedisct1/libsodium, commit
77e1ce5d6dee871c49ef211222ba18ef0c486bda (1.0.22-RELEASE)

Only a subset of all libsodium files are imported. They have been stripped of
unnecessary code and prefixed with `libsodium_`.

Original file path                                             -> New file path  
`src/libsodium/include/sodium/private/common.h`                -> `inc/libsodium_common.h`  
`src/libsodium/crypto_core/ed25519/ref10/fe_25_5/constants.h`  -> `inc/libsodium_constants.h`  
`src/libsodium/include/sodium/crypto_core_ed25519.h`           -> `inc/libsodium_crypto_core_ed25519.h`  
`src/libsodium/include/sodium/private/ed25519_ref10.h`         -> `inc/libsodium_ed25519_ref10.h`  
`src/libsodium/include/sodium/private/ed25519_ref10_fe_25_5.h` -> `inc/libsodium_ed25519_ref10_fe_25_5.h`  
`src/libsodium/include/sodium/export.h`                        -> `inc/libsodium_export.h`  
`src/libsodium/crypto_core/ed25519/ref10/fe_25_5/fe.h`         -> `inc/libsodium_fe.h`  
`src/libsodium/include/sodium/utils.h`                         -> `inc/libsodium_utils.h`  
`src/libsodium/crypto_core/ed25519/core_ed25519.c`             -> `src/libsodium_core_ed25519.c`  
`src/libsodium/crypto_core/ed25519/ref10/ed25519_ref10.c`      -> `src/libsodium_ed25519_ref10.c`  
`src/libsodium/sodium/utils.c`                                 -> `src/libsodium_utils.c`  

Licensed under the ISC license, see [LICENSE](LICENSE) for full license text.
