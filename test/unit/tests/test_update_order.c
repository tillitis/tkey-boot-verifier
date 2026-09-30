// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

// clang-format off
#include <stdarg.h>
#include <stddef.h>
#include <setjmp.h>
#include <stdint.h>
#include <cmocka.h> // cmocka need to be included last
// clang-format on

#include <string.h>

#include "../../verifier/update.h"

static void test_update_should_call_sys_preload_delete_first(void **state);

int main(void)
{
	const struct CMUnitTest tests[] = {
	    cmocka_unit_test(test_update_should_call_sys_preload_delete_first),
	};

	return cmocka_run_group_tests(tests, NULL, NULL);
}

static void test_update_should_call_sys_preload_delete_first(void **state)
{
	const size_t app_size = 200;
	const size_t write_size = 127;
	uint8_t app_buf[1000] = {0};
	_Static_assert(app_size < 2 * write_size,
		       "We assume app fits in two writes");
	_Static_assert(sizeof(app_buf) >= 2 * write_size,
		       "We assume the buffer holds enough data for two writes");
	memset(app_buf, 0x55, app_size);
	uint8_t app_digest[32] = {0};
	uint8_t app_signature[64] = {0};
	struct update_ctx ctx = {0};

	expect_function_call(__wrap_sys_preload_delete);
	expect_function_call(__wrap_sys_preload_store);
	expect_function_call(__wrap_sys_preload_store);
	expect_function_call(__wrap_sys_preload_store_fin);

	update_init(&ctx, app_size, app_digest, app_signature);
	update_write(&ctx, app_buf, write_size);
	update_write(&ctx, app_buf + write_size, write_size);
	update_finalize(&ctx);
}

int __wrap_sys_preload_delete(void)
{
	function_called();

	return 0;
}

int __wrap_sys_preload_store(uint32_t offset, void *app, size_t len)
{
	function_called();

	return 0;
}

int __wrap_sys_preload_store_fin(size_t len, uint8_t digest[32],
				 uint8_t signature[64])
{
	function_called();

	return 0;
}
