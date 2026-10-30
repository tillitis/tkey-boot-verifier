// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef MAC_H
#define MAC_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <tkey/syscall.h>

void mac_calc_key(const uint8_t key[32], const struct app_metadata *metadata,
		  uint8_t out_mac[32]);
void mac_calc(const struct app_metadata *metadata, uint8_t out_mac[32]);
bool mac_verify_key(const uint8_t key[32], const struct app_metadata *metadata);
bool mac_verify(const struct app_metadata *metadata);

#endif
