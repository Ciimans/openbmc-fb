/**
 * Copyright 2020-present Facebook. All Rights Reserved.
 *
 * This program file is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the
 * Free Software Foundation; version 2 of the License.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License
 * for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program in a file named COPYING; if not, write to the
 * Free Software Foundation, Inc.,
 * 51 Franklin Street, Fifth Floor,
 * Boston, MA 02110-1301 USA
 */

// constants/examples used for testing
package tests

import (
	"github.com/facebook/openbmc/tools/flashy/lib/utils"
)

const ExampleWedge100MemInfo = `MemTotal:         246900 kB
MemFree:           88928 kB
MemAvailable:     192496 kB
Buffers:               0 kB
Cached:           107288 kB
SwapCached:            0 kB
Active:            62848 kB
Inactive:          80988 kB
Active(anon):      36872 kB
Inactive(anon):      340 kB
Active(file):      25976 kB
Inactive(file):    80648 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:             0 kB
SwapFree:              0 kB
Dirty:                 0 kB
Writeback:             0 kB
AnonPages:         36560 kB
Mapped:            12680 kB
Shmem:               664 kB
Slab:               8800 kB
SReclaimable:       3440 kB
SUnreclaim:         5360 kB
KernelStack:         656 kB
PageTables:         1184 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:      123448 kB
Committed_AS:     248268 kB
VmallocTotal:     770048 kB
VmallocUsed:        1280 kB
VmallocChunk:     635736 kB`

// minilaketb healthd-config.json
const ExampleMinilaketbHealthdConfigJSON = `{
	"version": "1.0",
	"heartbeat": {
		"interval": 500
	},
	"bmc_cpu_utilization": {
		"enabled": true,
		"window_size": 120,
		"monitor_interval": 1,
		"threshold": [
			{
				"value": 80.0,
				"hysteresis": 5.0,
				"action": [
					"log-warning",
					"bmc-error-trigger"
				]
			}
		]
	},
	"bmc_mem_utilization": {
		"enabled": true,
		"enable_panic_on_oom": true,
		"window_size": 120,
		"monitor_interval": 1,
		"threshold": [
			{
				"value": 60.0,
				"hysteresis": 5.0,
				"action": [
					"log-warning"
				]
			},
			{
				"value": 70.0,
				"hysteresis": 5.0,
				"action": [
					"log-critical",
					"bmc-error-trigger"
				]
			},
			{
				"value": 95.0,
				"hysteresis": 5.0,
				"action": [
					"log-critical",
					"bmc-error-trigger",
					"reboot"
				]
			}
		]
	},
	"i2c": {
		"enabled": false,
		"busses": [
			0,
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
			9,
			10,
			11,
			12,
			13
		]
	},
	"ecc_monitoring": {
		"enabled": false,
		"ecc_address_log": false,
		"monitor_interval": 2,
		"recov_max_counter": 255,
		"unrec_max_counter": 15,
		"recov_threshold": [
			{
				"value": 0.0,
				"action": [
					"log-critical",
					"bmc-error-trigger"
				]
			},
			{
				"value": 50.0,
				"action": [
					"log-critical"
				]
			},
			{
				"value": 90.0,
				"action": [
					"log-critical"
				]
			}
		],
		"unrec_threshold": [
			{
				"value": 0.0,
				"action": [
					"log-critical",
					"bmc-error-trigger"
				]
			},
			{
				"value": 50.0,
				"action": [
					"log-critical"
				]
			},
			{
				"value": 90.0,
				"action": [
					"log-critical"
				]
			}
		]
	},
	"bmc_health": {
		"enabled": false,
		"monitor_interval": 2,
		"regenerating_interval": 1200
	},
	"verified_boot": {
		"enabled": false
	}
}`

// minimal part of minilaketb json
const ExampleMinimalHealthdConfigJSON = `{
	"bmc_mem_utilization": {
		"threshold": [
			{
				"value": 60.0,
				"hysteresis": 5.0,
				"action": [
					"log-warning"
				]
			},
			{
				"value": 70.0,
				"hysteresis": 5.0,
				"action": [
					"log-critical",
					"bmc-error-trigger"
				]
			},
			{
				"value": 95.0,
				"hysteresis": 5.0,
				"action": [
					"log-critical",
					"bmc-error-trigger",
					"reboot"
				]
			}
		]
	}
}`

const ExampleCorruptedHealthdConfig = `{
	"BMC_MEM_UTILIZATI: {
		"threshold": [
			{
				"value": 60.0,
				"hysteresis": 5.0,
				"action": [
					"log-warning"
				]
			},
			{
				"value": 70.0,
				"hysteresis": 5.0,
				"action": [
					"log-critical",
					"bmc-error-trigger"
				]
			},
			{
				"value": 95.0,
				"hysteresis": 5.0,
				"action": [
					"log-critical",
					"bmc-error-trigger",
					"reboot"
				]
			}
		]
	}
}`

var ExampleMinilaketbHealthdConfig = utils.HealthdConfig{
	utils.BmcMemUtil{
		[]utils.BmcMemThres{
			utils.BmcMemThres{
				Value:   float32(60.0),
				Actions: []string{"log-warning"},
			},
			utils.BmcMemThres{
				Value:   float32(70.0),
				Actions: []string{"log-critical", "bmc-error-trigger"},
			},
			utils.BmcMemThres{
				Value:   float32(95.0),
				Actions: []string{"log-critical", "bmc-error-trigger", "reboot"},
			},
		},
	},
}

var ExampleNoRebootHealthdConfig = utils.HealthdConfig{
	utils.BmcMemUtil{
		[]utils.BmcMemThres{
			utils.BmcMemThres{
				Value:   float32(60.0),
				Actions: []string{"log-warning"},
			},
			utils.BmcMemThres{
				Value:   float32(70.0),
				Actions: []string{"log-critical", "bmc-error-trigger"},
			},
		},
	},
}

const ExampleWedge100ProcMtdFile = `dev:    size   erasesize  name
mtd0: 00060000 00010000 "u-boot"
mtd1: 00020000 00010000 "env"
mtd2: 00400000 00010000 "kernel"
mtd3: 01780000 00010000 "rootfs"
mtd4: 00400000 00010000 "data0"
mtd5: 02000000 00010000 "flash0"`

const ExampleWedge100ProcMountsFile = `rootfs / rootfs rw 0 0
proc /proc proc rw,relatime 0 0
sysfs /sys sysfs rw,relatime 0 0
devtmpfs /dev devtmpfs rw,relatime,size=117120k,nr_inodes=29280,mode=755 0 0
tmpfs /run tmpfs rw,nosuid,nodev,mode=755 0 0
tmpfs /var/volatile tmpfs rw,relatime 0 0
/dev/mtdblock4 /mnt/data jffs2 rw,relatime 0 0
devpts /dev/pts devpts rw,relatime,gid=5,mode=620 0 0`

const ExampleTiogapass1ProcMtdFile = `dev:    size   erasesize  name
mtd0: 00060000 00010000 "romx"
mtd1: 00020000 00010000 "env"
mtd2: 00060000 00010000 "u-boot"
mtd3: 01b20000 00010000 "fit"
mtd4: 00400000 00010000 "data0"
mtd5: 02000000 00010000 "flash1"
mtd6: 01ff0000 00010000 "flash1rw"
mtd7: 00060000 00010000 "rom"
mtd8: 00020000 00010000 "envro"
mtd9: 00060000 00010000 "u-bootro"
mtd10: 01b20000 00010000 "fitro"
mtd11: 00400000 00010000 "dataro"
mtd12: 02000000 00010000 "flash0"`

const ExampleTiogapass1VbootUtilFile = `ROM executed from:       0x00000058
ROM KEK certificates:    0x00011544
ROM handoff marker:      0x00000000
U-Boot executed from:    0x28084350
U-Boot certificates:     0x2808054c
Certificates fallback:   Sat Jun 17 13:17:06 2017
Certificates time:       Sat Jun 17 13:17:06 2017
U-Boot fallback:         Wed Oct 30 18:54:54 2019
U-Boot time:             Thu Feb 27 15:37:05 2020
Kernel fallback:         not set
Kernel time:             not set
Flags force_recovery:    0x00
Flags hardware_enforce:  0x01
Flags software_enforce:  0x01
Flags recovery_boot:     0x00
Flags recovery_retried:  0x00

ROM version:             fbtp-v7.2
ROM U-Boot version:      2016.07

Status CRC: 0x76d9
TPM status  (0)
Status type (0) code (0)
OpenBMC was verified correctly`

// strings -n30 flash-wedge100 | grep U-Boot
const ExampleWedge100ImageUbootStrings = `U-Boot executed from:    0x%08x
U-Boot certificates:     0x%08x
MU-Boot EFI: Relocation at %p is out of range (%lx)
U-Boot 2016.07 wedge100-v2020.07.1 (Feb 12 2020 - 18:32:20 +0000)
U-Boot fitImage for Facebook OpenBMC/1.0/wedge100`

// strings -n30 flash-tiogapass1 | grep U-Boot
const ExampleTiogaPass1ImageUbootStrings = `U-Boot SPL 2016.07 fbtp-v2020.09.1 (Feb 27 2020 - 23:35:02)
U-Boot configuration was not verified.
U-Boot executed from:    0x%08x
U-Boot certificates:     0x%08x
No valid device tree binary found - please append one to U-Boot binary, use u-boot-dtb.bin or define CONFIG_OF_EMBED. For sandbox, use -d <file.dtb>
U-Boot 2016.07 fbtp-v2020.09.1 (Feb 27 2020 - 23:35:07 +0000)
U-Boot fitImage for Facebook OpenBMC/1.0/fbtp
U-Boot executed from:    0x%08x
U-Boot certificates:     0x%08x
No valid device tree binary found - please append one to U-Boot binary, use u-boot-dtb.bin or define CONFIG_OF_EMBED. For sandbox, use -d <file.dtb>
U-Boot 2016.07 fbtp-v2020.09.1 (Feb 27 2020 - 23:35:02 +0000)
U-Boot fitImage for Facebook OpenBMC/1.0/fbtp`
