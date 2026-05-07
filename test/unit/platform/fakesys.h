// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef FAKESYS_H
#define FAKESYS_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

void fakesys_preload_set_metadata(uint8_t digest[32], uint8_t signature[64],
				  uint8_t pubkey[32]);
void fakesys_preload_erase(void);
bool fakesys_preload_range_contains_ff(uint32_t start, uint32_t stop);
bool fakesys_preload_range_contains_data(uint32_t offset, void *data,
					 size_t data_len);

#endif
