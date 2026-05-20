# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

import struct
from base64 import b64decode
from dataclasses import dataclass
from os import PathLike
from pathlib import Path

PUBKEY_LEN = 32


@dataclass
class PublicKeyInfo:
    alg: bytes
    key_num: bytes
    key: bytes
    comment: str


def parse_pubkey_text(text: str) -> PublicKeyInfo:
    lines = text.splitlines()
    comment = lines[0]

    packed_key_info = b64decode(lines[1], validate=True)
    alg, key_num, key = struct.unpack("2s8s32s", packed_key_info)

    info = PublicKeyInfo(alg=alg, key_num=key_num, key=key, comment=comment)

    return info


def parse_pubkey_file(path: str | PathLike[str]) -> PublicKeyInfo:
    text = Path(path).read_text(encoding="utf-8")

    return parse_pubkey_text(text)
