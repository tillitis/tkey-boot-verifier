// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef VERIFY_H
#define VERIFY_H

#include <stdbool.h>
#include <stdint.h>
#include <tkey/syscall.h>

int reset_if_verified(const uint8_t pubkey[32], const uint8_t app_name[16],
		      enum reset_start reset_type, const uint8_t app_digest[32],
		      const uint8_t app_signature[64], bool installable);

#endif
