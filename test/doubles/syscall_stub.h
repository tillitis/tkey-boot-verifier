// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef SYSCALL_STUB_H
#define SYSCALL_STUB_H

#include <stdint.h>

void sys_stub_set_digsig(uint8_t digest[32], uint8_t signature[64]);

#endif
