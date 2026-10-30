// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef FAKEMAC_H
#define FAKEMAC_H

#include <stdbool.h>
#include <stdint.h>
#include <tkey/syscall.h>

void mac_calc(struct app_metadata *metadata, uint8_t out_mac[32]);
bool mac_verify(struct app_metadata *metadata);

#endif
