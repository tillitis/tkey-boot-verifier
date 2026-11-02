// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

// clang-format off
#include <stdarg.h>
#include <stddef.h>
#include <setjmp.h>
#include <stdint.h>
#include <cmocka.h> // cmocka need to be included last
// clang-format on

#include "../../verifier/app_proto.h"
#include "../../verifier/pubkey.h"
#include "../platform/fakesys.h"
#include <string.h>
#include <tkey/proto.h>

// Private key used to sign:
//   000102030405060708090a0b0c0d0e0f
//   000102030405060708090a0b0c0d0e0f
//   9b62773323ef41a11834824194e55164
//   d325eb9cdcc10ddda7d10ade4fbd8f6d
static const uint8_t PUBKEY[32] = {
    0x9b, 0x62, 0x77, 0x33, 0x23, 0xef, 0x41, 0xa1, 0x18, 0x34, 0x82,
    0x41, 0x94, 0xe5, 0x51, 0x64, 0xd3, 0x25, 0xeb, 0x9c, 0xdc, 0xc1,
    0x0d, 0xdd, 0xa7, 0xd1, 0x0a, 0xde, 0x4f, 0xbd, 0x8f, 0x6d,
};

static const uint8_t VALID_PUBKEY[32] = {
    0xA4, 0xF3, 0xBC, 0x6B, 0x4D, 0xFA, 0x4A, 0x81, 0x36, 0xA2, 0xC4,
    0x2E, 0x61, 0x23, 0xC0, 0x9C, 0x81, 0xCC, 0x1C, 0x77, 0xE6, 0x7B,
    0x1E, 0xE2, 0x07, 0x3C, 0x37, 0xB3, 0xA4, 0x44, 0x44, 0xA3,
};
static const uint8_t INVALID_PUBKEY[32] = {
    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
};

static const uint8_t SIG_VALID_PUBKEY[64] = {
    0x9f, 0x75, 0x55, 0x60, 0x92, 0x64, 0x15, 0xcb, 0x86, 0xd3, 0x7e,
    0xb6, 0x15, 0xe4, 0xc5, 0x1c, 0x80, 0x70, 0x1f, 0x9f, 0x3e, 0x3a,
    0x40, 0xb4, 0xec, 0x03, 0x1c, 0x24, 0xa3, 0x48, 0xf8, 0xe6, 0x18,
    0xee, 0xb4, 0xc7, 0xcd, 0x94, 0xb1, 0x1a, 0x90, 0x10, 0x09, 0x2c,
    0x54, 0xab, 0xba, 0xe5, 0x52, 0x0e, 0x99, 0x15, 0xcf, 0x34, 0x7e,
    0x37, 0x30, 0x3f, 0xd5, 0x6d, 0x69, 0x8a, 0xfb, 0x05,
};

static const uint8_t SIG_INVALID_PUBKEY[64] = {
    0x98, 0x58, 0x5e, 0x7f, 0xd5, 0x58, 0x20, 0xf3, 0xad, 0x28, 0x76,
    0x78, 0xc1, 0xfa, 0x3b, 0x07, 0x43, 0x4c, 0x70, 0x0b, 0x69, 0x8f,
    0xcf, 0xcd, 0x82, 0xe7, 0x92, 0x7a, 0xb2, 0x33, 0x4e, 0x29, 0x73,
    0xdf, 0xf8, 0x74, 0x70, 0xa4, 0x10, 0x1a, 0x99, 0x6d, 0x96, 0x8e,
    0xc8, 0x31, 0xd8, 0x2f, 0xce, 0x64, 0x4e, 0xd5, 0xd8, 0x33, 0xd9,
    0xf5, 0x5f, 0xca, 0x1b, 0x62, 0xda, 0xd4, 0x77, 0x07,
};

static int response = -1;

void reset_reponse()
{
	response = -1;
}

static void test_reject_low_order_pubkey(void **state)
{
	// Arrange
	reset_reponse();

	struct app_metadata metadata = {0};
	memcpy(metadata.pubkey, PUBKEY, sizeof(PUBKEY));

	fakesys_preload_set_metadata(&metadata);

	struct packet pkt = {0};
	pkt.hdr.len = 128;
	// new pubkey pkt.cmd[1]
	// signature pkt.cmd[33]
	memcpy(&pkt.cmd[1], INVALID_PUBKEY, sizeof(INVALID_PUBKEY));
	memcpy(&pkt.cmd[33], SIG_INVALID_PUBKEY, sizeof(SIG_INVALID_PUBKEY));

	// Act and assert...
	store_pubkey(pkt);
	assert_int_equal(response, FRAME_STATUS_NOK);
}

static void test_valid_pubkey(void **state)
{
	// Arrange
	reset_reponse();

	struct app_metadata metadata = {0};
	memcpy(metadata.pubkey, PUBKEY, sizeof(PUBKEY));

	fakesys_preload_set_metadata(&metadata);

	struct packet pkt = {0};
	pkt.hdr.len = 128;
	// new pubkey pkt.cmd[1]
	// signature pkt.cmd[33]
	memcpy(&pkt.cmd[1], VALID_PUBKEY, sizeof(VALID_PUBKEY));
	memcpy(&pkt.cmd[33], SIG_VALID_PUBKEY, sizeof(SIG_VALID_PUBKEY));

	expect_function_call(__wrap_sys_preload_set_metadata);

	// Act and assert...
	store_pubkey(pkt);
	assert_int_equal(response, FRAME_STATUS_OK);
}

static void test_valid_pubkey_invalid_sig(void **state)
{
	// Arrange
	reset_reponse();


	struct app_metadata metadata = {0};
	memcpy(metadata.pubkey, PUBKEY, sizeof(PUBKEY));

	fakesys_preload_set_metadata(&metadata);

	struct packet pkt = {0};
	pkt.hdr.len = 128;
	// new pubkey pkt.cmd[1]
	// signature pkt.cmd[33]
	memcpy(&pkt.cmd[1], VALID_PUBKEY, sizeof(VALID_PUBKEY));
	memcpy(&pkt.cmd[33], SIG_INVALID_PUBKEY, sizeof(SIG_INVALID_PUBKEY));

	// Act and assert...
	store_pubkey(pkt);
	assert_int_equal(response, FRAME_STATUS_NOK);
}

int __wrap_sys_preload_set_metadata(struct app_metadata *metadata)
{
	function_called();

	return 0;
}

bool __wrap_user_is_present(int c)
{
	return true;
}

void __wrap_appreply(struct frame_header hdr, enum appcmd rspcode, void *buf)
{
	response = ((uint8_t *)buf)[0];
}

int main(void)
{
	const struct CMUnitTest tests[] = {
	    cmocka_unit_test(test_reject_low_order_pubkey),
	    cmocka_unit_test(test_valid_pubkey),
	    cmocka_unit_test(test_valid_pubkey_invalid_sig),
	};

	return cmocka_run_group_tests(tests, NULL, NULL);
}
