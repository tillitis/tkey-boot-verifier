# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

from .drivers.testapp_probe import TestappProbe
from .drivers.tkey_mgt import TKeyMgt

testapp_probe = TestappProbe("../../testapp-probe")
tkey_mgt = TKeyMgt("../../tkey-mgt")
