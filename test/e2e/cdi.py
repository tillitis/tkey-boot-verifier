# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

from hashlib import blake2s
from pathlib import Path

from .pubkey import parse_pubkey_file
from .tt.models.cdi import DEV_UDS, UDS_LEN, CdiDomain, calc_cdi_castor_direct

PUBKEY_LEN = 32

calc_cdi_castor_measured = calc_cdi_castor_direct


def calc_cdi_castor_verified(
    verifier_bin: bytes, vendor_pubkey: bytes, uds: bytes = DEV_UDS
) -> bytes:
    """Calculate CDI of verified app booted via tkey-boot-verifier onto TKey Castor."""
    assert len(uds) == UDS_LEN
    assert len(vendor_pubkey) == PUBKEY_LEN

    # Firmware boots verifier from flash slot 0
    verifier_cdi = calc_cdi_castor_measured(verifier_bin, uds)

    # Verifier calculates measured_id seed
    measured_id_seed = blake2s(vendor_pubkey).digest()

    # Firmware calculates measured_id
    measured_id = blake2s(measured_id_seed, key=verifier_cdi).digest()

    # Firmware boots verified app from slot 1
    verifiee_domain = bytes([CdiDomain.UseMeasuredId])
    verifiee_cdi_hash = blake2s(key=uds)
    verifiee_cdi_hash.update(verifiee_domain)
    verifiee_cdi_hash.update(measured_id)

    return verifiee_cdi_hash.digest()


def format_cdi(cdi: bytes) -> str:
    return cdi.hex().zfill(PUBKEY_LEN * 2)


def expected_cdi_verified(verifier_path: str, pubkey_path: str) -> str:
    return format_cdi(
        calc_cdi_castor_verified(
            Path(verifier_path).read_bytes(),
            parse_pubkey_file(pubkey_path).key,
        )
    )


def expected_cdi_measured(testapp_path: str) -> str:
    return format_cdi(calc_cdi_castor_measured(Path(testapp_path).read_bytes()))
