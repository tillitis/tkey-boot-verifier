// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef VERIFY_H
#define VERIFY_H

#include <stdint.h>
#include <tkey/syscall.h>

int reset_if_verified(uint8_t pubkey[32], enum reset_start reset_type,
		      uint8_t app_digest[32], uint8_t app_signature[64]);

#endif
