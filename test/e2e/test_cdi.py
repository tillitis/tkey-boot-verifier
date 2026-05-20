# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

from .assets import pubkey0_path, pubkey1_path, testapp_a, testapp_b, verifier_path
from .cdi import expected_cdi_measured, expected_cdi_verified
from .commands import testapp_probe, tkey_mgt, tkey_runapp
from .drivers.testapp_probe import FwResetType, VerifierResetDst
from .tt.drivers.tkey import TKey

# There is currently no way to specify "no verifier reset destination". Reusing
# APP1 and calling it NO_DST to signal test intent.
NO_DST = VerifierResetDst.APP1


def test_verified_app_from_flash_should_get_expected_cdi_on_power_on(
    tkey: TKey,
) -> None:
    assert testapp_probe.get_nameversion(tkey) == testapp_a.nameversion
    assert testapp_probe.get_cdi(tkey) == expected_cdi_verified(
        verifier_path, pubkey0_path
    )


def test_app_cdi_should_be_unchanged_after_app_update(
    tkey: TKey,
) -> None:
    assert testapp_probe.get_nameversion(tkey) == testapp_a.nameversion
    first_app_cdi = testapp_probe.get_cdi(tkey)

    tkey_mgt.install_app(tkey, testapp_b.path, testapp_b.sig0_path)

    assert testapp_probe.get_nameversion(tkey) == testapp_b.nameversion
    assert testapp_probe.get_cdi(tkey) == first_app_cdi


def test_app_cdi_should_change_when_pubkey_is_changed(
    tkey: TKey,
) -> None:
    testapp = testapp_a

    assert testapp_probe.get_nameversion(tkey) == testapp.nameversion
    cdi_before_install = testapp_probe.get_cdi(tkey)
    assert cdi_before_install == expected_cdi_verified(verifier_path, pubkey0_path)

    tkey_mgt.install_pubkey(tkey, pubkey1_path)
    tkey_mgt.install_app(tkey, testapp.path, testapp.sig1_path)

    assert testapp_probe.get_nameversion(tkey) == testapp.nameversion
    cdi_after_install = testapp_probe.get_cdi(tkey)
    assert cdi_after_install == expected_cdi_verified(verifier_path, pubkey1_path)
    assert cdi_after_install != cdi_before_install


def test_verified_app_from_flash_should_get_expected_cdi_after_reset(
    tkey: TKey,
) -> None:
    testapp_probe.reset(tkey, FwResetType.START_FLASH0, VerifierResetDst.APP1)

    assert testapp_probe.get_nameversion(tkey) == testapp_a.nameversion
    assert testapp_probe.get_cdi(tkey) == expected_cdi_verified(
        verifier_path, pubkey0_path
    )


def test_app_loaded_directly_from_flash_slot_1_should_get_expected_cdi(
    tkey: TKey,
) -> None:
    testapp_probe.reset(tkey, FwResetType.START_FLASH1, NO_DST)

    assert testapp_probe.get_nameversion(tkey) == testapp_a.nameversion
    assert testapp_probe.get_cdi(tkey) == expected_cdi_measured(testapp_a.path)


def test_app_loaded_directly_from_client_should_get_expected_cdi(
    tkey: TKey,
) -> None:
    testapp_probe.reset(tkey, FwResetType.START_CLIENT, NO_DST)
    tkey_runapp.load(tkey, testapp_b.path)

    assert testapp_probe.get_nameversion(tkey) == testapp_b.nameversion
    assert testapp_probe.get_cdi(tkey) == expected_cdi_measured(testapp_b.path)


def test_app_and_verifier_from_client_should_get_expected_cdi(tkey: TKey) -> None:
    # With pubkey0
    tkey_mgt.boot(tkey, testapp_b.path, testapp_b.sig0_path, pubkey0_path)

    assert testapp_probe.get_nameversion(tkey) == testapp_b.nameversion
    assert testapp_probe.get_cdi(tkey) == expected_cdi_verified(
        verifier_path, pubkey0_path
    )

    # With pubkey1
    tkey_mgt.boot(tkey, testapp_b.path, testapp_b.sig1_path, pubkey1_path)

    assert testapp_probe.get_nameversion(tkey) == testapp_b.nameversion
    assert testapp_probe.get_cdi(tkey) == expected_cdi_verified(
        verifier_path, pubkey1_path
    )
