// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef APP_UPDATE_H
#define APP_UPDATE_H

#include <stddef.h>
#include <stdint.h>

int write_app(uint32_t addr, uint8_t *data, size_t sz);

#endif
