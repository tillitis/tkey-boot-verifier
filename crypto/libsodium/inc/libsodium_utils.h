#ifndef sodium_utils_H
#define sodium_utils_H

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "libsodium_export.h"

int sodium_is_zero(const unsigned char *n, const size_t nlen);

#endif
