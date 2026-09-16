# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

from pathlib import Path

from .tt.drivers.qemu_tkey import QEmuTKey, QEmuTKeyMachine
from .tt.drivers.tkey import TKey, TKeyType, Udi


def new_qemu_tkey(
    tkey_type: TKeyType,
    qemu_path: Path,
    qemu_usb_mux_path: Path,
    firmware_path: Path,
    flash_image_path: Path,
) -> TKey:
    match tkey_type:
        case TKeyType.CastorPre:
            return QEmuTKey(
                tkey_type,
                QEmuTKeyMachine.TK1_CASTOR,
                qemu_path,
                firmware_path,
                flash_image_path,
                qemu_usb_mux_path,
                Udi(vid=0x1337, pid=3, rev=0, serial=0x012345),
            )
        case _:
            raise ValueError(f"Unsupported type: {tkey_type}")
