// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

// clang-format off
#include <stdarg.h>
#include <stddef.h>
#include <stdbool.h>
#include <setjmp.h>
#include <stdint.h>
#include <cmocka.h> // cmocka need to be included last
// clang-format on

#include <tkey/syscall.h>

#include "fakemac.h"

// Calculates a MAC over the metadata, returns the result in out_mac.
void mac_calc(struct app_metadata *metadata, uint8_t out_mac[32])
{
	return;
}

// Returns true if MAC verifies
bool mac_verify(struct app_metadata *metadata)
{
	return true;
}
