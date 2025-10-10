// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef FAKESYS_H
#define FAKESYS_H

#include <stdint.h>

void fakesys_set_digsig(uint8_t digest[32], uint8_t signature[64]);

#endif
