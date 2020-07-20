#!/bin/sh
#
# Copyright 2019-present Facebook. All Rights Reserved.
#
# This program file is free software; you can redistribute it and/or modify it
# under the terms of the GNU General Public License as published by the
# Free Software Foundation; version 2 of the License.
#
# This program is distributed in the hope that it will be useful, but WITHOUT
# ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
# FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License
# for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program in a file named COPYING; if not, write to the
# Free Software Foundation, Inc.,
# 51 Franklin Street, Fifth Floor,
# Boston, MA 02110-1301 USA

PATH=/sbin:/bin:/usr/sbin:/usr/bin

. /usr/local/fbpackages/utils/ast-functions

# CHKPT_COMPLETE

BOARD_ID=$(get_bmc_board_id)
RCVY_CNT=0x0
PANIC_CNT=0x0
LAST_RCVY_ERROR=0x0
LAST_PANIC_ERROR=0x0
MAJOR_ERROR=0x0
MINOR_ERROR=0x0
LAST_RCVY_ERROR_MSG=""
LAST_PANIC_ERROR_MSG=""
MAJOR_ERROR_MSG=""
MINOR_ERROR_MSG=""
ret=0
# NIC EXP BMC
if [ $BOARD_ID -eq 9 ]; then
    for i in {1..3}; do
        /usr/sbin/i2craw 9 0x38 -w "0x0F 0x09" > /dev/null
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        RCVY_CNT=$(/usr/sbin/i2cget -y 9 0x38 0x04)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        LAST_RCVY_ERROR=$(/usr/sbin/i2cget -y 9 0x38 0x05)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        PANIC_CNT=$(/usr/sbin/i2cget -y 9 0x38 0x06)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        LAST_PANIC_ERROR=$(/usr/sbin/i2cget -y 9 0x38 0x07)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        MAJOR_ERROR=$(/usr/sbin/i2cget -y 9 0x38 0x08)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        MINOR_ERROR=$(/usr/sbin/i2cget -y 9 0x38 0x09)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

# Baseboard BMC
elif [ $BOARD_ID -eq 14 ] || [ $BOARD_ID -eq 7 ]; then
    for i in {1..3}; do
        /usr/sbin/i2craw 12 0x38 -w "0x0F 0x09" > /dev/null
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        RCVY_CNT=$(/usr/sbin/i2cget -y 12 0x38 0x04)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        LAST_RCVY_ERROR=$(/usr/sbin/i2cget -y 12 0x38 0x05)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        PANIC_CNT=$(/usr/sbin/i2cget -y 12 0x38 0x06)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        LAST_PANIC_ERROR=$(/usr/sbin/i2cget -y 12 0x38 0x07)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        MAJOR_ERROR=$(/usr/sbin/i2cget -y 12 0x38 0x08)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done

    for i in {1..3}; do
        MINOR_ERROR=$(/usr/sbin/i2cget -y 12 0x38 0x09)
        ret=$?
        if [[ "$ret" == "0" ]]; then
            break
        fi
    done
else
  echo "Is board id correct(id=$BOARD_ID)?"
fi

echo "Check PFR MailBox Status..."
if [[ "$MAJOR_ERROR" -ne "0x0" ]] || [[ "$MINOR_ERROR" -ne "0x0" ]]; then
    case $MAJOR_ERROR in
        0x01 )
            MAJOR_ERROR_MSG="MAJOR_ERROR_BMC_AUTH_FAILED"
            case $MINOR_ERROR in
                0x01 )
                    MINOR_ERROR_MSG="MINOR_ERROR_AUTH_ACTIVE"
                ;;
                0x02 )
                    MINOR_ERROR_MSG="MINOR_ERROR_AUTH_RECOVERY"
                ;;
                0x03 )
                    MINOR_ERROR_MSG="MINOR_ERROR_AUTH_ACTIVE_AND_RECOVERY"
                ;;
                0x04 )
                    MINOR_ERROR_MSG="MINOR_ERROR_AUTH_ALL_REGIONS"
                ;;
                0x06 )
                    MINOR_ERROR_MSG="MINOR_ERROR_CPLD_UPDATE_INVALID_SVN"
                ;;
                0x07 )
                    MINOR_ERROR_MSG="MINOR_ERROR_CPLD_UPDATE_AUTH_FAILED"
                ;;
                0x08 )
                    MINOR_ERROR_MSG="MINOR_ERROR_CPLD_UPDATE_EXCEEDED_MAX_FAILED_ATTEMPTS"
                ;;
                *)
                    MINOR_ERROR_MSG="unknown minor error"
                ;;
            esac
        ;;
        0x04 )
            MAJOR_ERROR_MSG="MAJOR_ERROR_UPDATE_FROM_BMC_FAILED"
            case $MINOR_ERROR in
                0x01 )
                    MINOR_ERROR_MSG="MINOR_ERROR_INVALID_UPDATE_INTENT"
                ;;
                0x02 )
                    MINOR_ERROR_MSG="MINOR_ERROR_FW_UPDATE_INVALID_SVN"
                ;;
                0x03 )
                    MINOR_ERROR_MSG="MINOR_ERROR_FW_UPDATE_AUTH_FAILED"
                ;;
                0x04 )
                    MINOR_ERROR_MSG="MINOR_ERROR_FW_UPDATE_EXCEEDED_MAX_FAILED_ATTEMPTS"
                ;;
                0x05 )
                    MINOR_ERROR_MSG="MINOR_ERROR_FW_UPDATE_ACTIVE_UPDATE_NOT_ALLOWED"
                ;;
                *)
                    MINOR_ERROR_MSG="unknown minor error"
                ;;
            esac
        ;;
        *)
            MAJOR_ERROR_MSG="unknown major error"
        ;;
    esac
    logger -t "pfr" -p daemon.crit "BMC, PFR - Major error: ${MAJOR_ERROR_MSG} (${MAJOR_ERROR}), Minor error: ${MINOR_ERROR_MSG} (${MINOR_ERROR})"
fi

if [[ "$RCVY_CNT" -ne "0x0" ]]; then
    logger -t "pfr" -p daemon.crit "BMC, PFR - Recovery count: $((RCVY_CNT))"
fi

if [[ "$LAST_RCVY_ERROR" -ne "0x0" ]]; then
    case $LAST_RCVY_ERROR in
        0x01 )
            LAST_RCVY_ERROR_MSG="LAST_RECOVERY_PCH_ACTIVE_FAIL"
        ;;
        0x02 )
            LAST_RCVY_ERROR_MSG="LAST_RECOVERY_PCH_RECOVERY_FAIL"
        ;;
        0x03 )
            LAST_RCVY_ERROR_MSG="LAST_RECOVERY_ME_LAUNCH_FAIL"
        ;;
        0x04 )
            LAST_RCVY_ERROR_MSG="LAST_RECOVERY_ACM_LAUNCH_FAIL"
        ;;
        0x05 )
            LAST_RCVY_ERROR_MSG="LAST_RECOVERY_IBB_LAUNCH_FAIL"
        ;;
        0x06 )
            LAST_RCVY_ERROR_MSG="LAST_RECOVERY_OBB_LAUNCH_FAIL"
        ;;
        0x07 )
            LAST_RCVY_ERROR_MSG="LAST_RECOVERY_BMC_ACTIVE_FAIL"
        ;;
        0x08 )
            LAST_RCVY_ERROR_MSG="LAST_RECOVERY_BMC_RECOVERY_FAIL"
        ;;
        0x09 )
            LAST_RCVY_ERROR_MSG="LAST_RECOVERY_BMC_LAUNCH_FAIL"
        ;;
        0x0a )
            LAST_RCVY_ERROR_MSG="LAST_RECOVERY_FORCED_ACTIVE_FW_RECOVERY"
        ;;
        *)
            LAST_RCVY_ERROR_MSG="unknown recovery reason"
        ;;
    esac
    logger -t "pfr" -p daemon.crit "BMC, PFR - Last recovery reason: ${LAST_RCVY_ERROR_MSG} (${LAST_RCVY_ERROR})"
fi

if [[ "$PANIC_CNT" -ne "0x0" ]]; then
    logger -t "pfr" -p daemon.crit "BMC, PFR - Panic event count: $((PANIC_CNT))"
fi

if [[ "$LAST_PANIC_ERROR" -ne "0x0" ]]; then
    case $LAST_PANIC_ERROR in
        0x01 )
            LAST_PANIC_ERROR_MSG="LAST_PANIC_PCH_UPDATE_INTENT"
        ;;
        0x02 )
            LAST_PANIC_ERROR_MSG="LAST_PANIC_BMC_UPDATE_INTENT"
        ;;
        0x03 )
            LAST_PANIC_ERROR_MSG="LAST_PANIC_BMC_RESET_DETECTED"
        ;;
        0x04 )
            LAST_PANIC_ERROR_MSG="LAST_PANIC_BMC_WDT_EXPIRED"
        ;;
        0x05 )
            LAST_PANIC_ERROR_MSG="LAST_PANIC_ME_WDT_EXPIRED"
        ;;
        0x06 )
            LAST_PANIC_ERROR_MSG="LAST_PANIC_ACM_WDT_EXPIRED"
        ;;
        0x07 )
            LAST_PANIC_ERROR_MSG="LAST_PANIC_IBB_WDT_EXPIRED"
        ;;
        0x08 )
            LAST_PANIC_ERROR_MSG="LAST_PANIC_OBB_WDT_EXPIRED"
        ;;
        0x09 )
            LAST_PANIC_ERROR_MSG="LAST_PANIC_ACM_IBB_OBB_AUTH_FAILED"
        ;;
        *)
            LAST_PANIC_ERROR_MSG="unknown panic reason"
        ;;
    esac
    logger -t "pfr" -p daemon.crit "BMC, PFR - Last panic reason: ${LAST_PANIC_ERROR_MSG} (${LAST_PANIC_ERROR})"
fi
