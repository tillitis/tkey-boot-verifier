#ifndef sodium_common_H
#define sodium_common_H 1

#include <stdint.h>

#define COMPILER_ASSERT(X) (void) sizeof(char[(X) ? 1 : -1])

#if !defined(__clang__) && !defined(__GNUC__)
# ifdef __attribute__
#  undef __attribute__
# endif
# define __attribute__(a)
#endif

#ifndef CRYPTO_ALIGN
#  define CRYPTO_ALIGN(x) __attribute__ ((aligned(x)))
#endif

#endif
