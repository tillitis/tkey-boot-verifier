#ifndef crypto_core_ed25519_H
#define crypto_core_ed25519_H

#include <stddef.h>
#include "libsodium_export.h"

int crypto_core_ed25519_is_valid_point(const unsigned char *p)
            __attribute__ ((nonnull));

#endif
