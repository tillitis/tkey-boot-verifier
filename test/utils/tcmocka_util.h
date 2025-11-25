// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef TCMOCKA_UTIL_H
#define TCMOCKA_UTIL_H

#define print_subtest(...)                                                     \
	print_message("[  SUBTEST ]  ");                                       \
	print_message(__VA_ARGS__)

#endif
