// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

// clang-format off
#include <stdarg.h>
#include <stddef.h>
#include <setjmp.h>
#include <stdint.h>
#include <string.h>
#include <cmocka.h> // cmocka need to be included last
// clang-format on

#include "../verifier/mac.h"
#include <tkey/syscall.h>

const uint8_t key[32] = {
    0x9b, 0x62, 0x77, 0x33, 0x23, 0xef, 0x41, 0xa1, 0x18, 0x34, 0x82,
    0x41, 0x94, 0xe5, 0x51, 0x64, 0xd3, 0x25, 0xeb, 0x9c, 0xdc, 0xc1,
    0x0d, 0xdd, 0xa7, 0xd1, 0x0a, 0xde, 0x4f, 0xbd, 0x8f, 0x6d,
};

const struct app_metadata metadata = {
    .size = 256,
    .digest = {0x01, 0x02, 0x03, 0x04},
    .signature = {0x01, 0x02, 0x03, 0x04},
    .pubkey = {0x01, 0x02, 0x03, 0x04},
};

const uint8_t expected[32] = {0xdc, 0x54, 0xe1, 0xa3, 0xe5, 0xd2, 0x5b, 0x87,
			      0x48, 0x20, 0x74, 0x5a, 0x91, 0x4f, 0x3f, 0x3e,
			      0xc8, 0x9f, 0xee, 0xfb, 0x5b, 0xbc, 0x36, 0x5c,
			      0x6f, 0x93, 0x8c, 0x0f, 0x9c, 0x9a, 0x42, 0xfe};

static void test_mac_known_vector(void **state)
{
	(void)state;

	uint8_t actual[32];
	mac_calc_key(key, &metadata, actual);

	assert_memory_equal(actual, expected, sizeof(actual));
}

static void test_mac_modified_metadata(void **state)
{
	(void)state;

	struct app_metadata changed;
	uint8_t actual[32];

	changed = metadata;
	changed.size ^= 1;
	mac_calc_key(key, &changed, actual);
	assert_memory_not_equal(actual, expected, sizeof(actual));

	changed = metadata;
	changed.digest[sizeof(changed.digest) - 1] ^= 1;
	mac_calc_key(key, &changed, actual);
	assert_memory_not_equal(actual, expected, sizeof(actual));

	changed = metadata;
	changed.signature[sizeof(changed.signature) - 1] ^= 1;
	mac_calc_key(key, &changed, actual);
	assert_memory_not_equal(actual, expected, sizeof(actual));

	changed = metadata;
	changed.pubkey[sizeof(changed.pubkey) - 1] ^= 1;
	mac_calc_key(key, &changed, actual);
	assert_memory_not_equal(actual, expected, sizeof(actual));
}

static void test_mac_modified_key(void **state)
{
	(void)state;

	uint8_t changed_key[sizeof(key)];
	uint8_t actual[32];

	for (size_t i = 0; i < sizeof(key); i++) {
		changed_key[i] = key[i];
	}

	changed_key[sizeof(changed_key) - 1] ^= 1;

	mac_calc_key(changed_key, &metadata, actual);
	assert_memory_not_equal(actual, expected, sizeof(actual));
}

static void test_mac_ignores_metadata_mac(void **state)
{
	(void)state;

	struct app_metadata changed = metadata;
	uint8_t actual[32];

	for (size_t i = 0; i < sizeof(changed.mac); i++) {
		changed.mac[i] ^= 0xff;
	}

	mac_calc_key(key, &changed, actual);

	assert_memory_equal(actual, expected, sizeof(actual));
}

static void test_mac_verify_valid(void **state)
{
	(void)state;

	struct app_metadata input = metadata;
	memcpy(input.mac, expected, sizeof(input.mac));

	assert_true(mac_verify_key(key, &input));
}

static void test_mac_verify_invalid_mac(void **state)
{
	(void)state;

	struct app_metadata input = metadata;
	memcpy(input.mac, expected, sizeof(input.mac));

	/* Check that every byte of the MAC is verified. */
	for (size_t i = 0; i < sizeof(input.mac); i++) {
		input.mac[i] ^= 1;
		assert_false(mac_verify_key(key, &input));
		input.mac[i] ^= 1;
	}
}

static void test_mac_verify_modified_metadata(void **state)
{
	(void)state;

	struct app_metadata original = metadata;
	memcpy(original.mac, expected, sizeof(original.mac));

	struct app_metadata input = original;
	input.size ^= 1;
	assert_false(mac_verify_key(key, &input));

	input = original;
	input.digest[sizeof(input.digest) - 1] ^= 1;
	assert_false(mac_verify_key(key, &input));

	input = original;
	input.signature[sizeof(input.signature) - 1] ^= 1;
	assert_false(mac_verify_key(key, &input));

	input = original;
	input.pubkey[sizeof(input.pubkey) - 1] ^= 1;
	assert_false(mac_verify_key(key, &input));
}

static void test_mac_verify_wrong_key(void **state)
{
	(void)state;

	struct app_metadata input = metadata;
	uint8_t wrong_key[sizeof(key)];

	memcpy(input.mac, expected, sizeof(input.mac));
	memcpy(wrong_key, key, sizeof(wrong_key));
	wrong_key[0] ^= 1;

	assert_false(mac_verify_key(wrong_key, &input));
}

int main(void)
{
	const struct CMUnitTest tests[] = {
	    cmocka_unit_test(test_mac_known_vector),
	    cmocka_unit_test(test_mac_modified_metadata),
	    cmocka_unit_test(test_mac_modified_key),
	    cmocka_unit_test(test_mac_ignores_metadata_mac),
	    cmocka_unit_test(test_mac_verify_valid),
	    cmocka_unit_test(test_mac_verify_invalid_mac),
	    cmocka_unit_test(test_mac_verify_modified_metadata),
	    cmocka_unit_test(test_mac_verify_wrong_key),
	};

	return cmocka_run_group_tests(tests, NULL, NULL);
}
