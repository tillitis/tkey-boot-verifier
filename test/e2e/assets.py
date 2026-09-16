# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

from dataclasses import dataclass


@dataclass
class Testapp:
    path: str  # Path to app binary
    sig0_path: str  # Path to signature signed with key corresponding to pubkey0
    sig1_path: str  # Path to signature signed with key corresponding to pubkey1
    nameversion: str

    __test__ = False


pubkey0_path = "../../testapp/pubkey"
pubkey1_path = "../../testapp/pubkey.1"
testapp_a = Testapp(
    path="../../testapp/app_a.bin",
    sig0_path="../../testapp/app_a.bin.sig",
    sig1_path="../../testapp/app_a.bin.sig.1",
    nameversion="tk1 appA 0",
)
testapp_b = Testapp(
    path="../../testapp/app_b.bin",
    sig0_path="../../testapp/app_b.bin.sig",
    sig1_path="../../testapp/app_b.bin.sig.1",
    nameversion="tk1 appB 0",
)
verifier_path = "../../verifier/app.bin"
