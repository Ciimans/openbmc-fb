/*
 *
 * Copyright 2015-present Facebook. All Rights Reserved.
 *
 * This file contains code to support IPMI2.0 Specificaton available @
 * http://www.intel.com/content/www/us/en/servers/ipmi/ipmi-specifications.html
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 675 Mass Ave, Cambridge, MA 02139, USA.
 */
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <unistd.h>
#include <math.h>
#include <fcntl.h>
#include <errno.h>
#include <syslog.h>
#include <pthread.h>
#include <string.h>
#include <openbmc/libgpio.h>
#include <openbmc/obmc-i2c.h>
#include "fby3_common.h"

const char *slot_usage = "slot1|slot2|slot3|slot4";
const char *slot_list[] = {"all", "slot1", "slot2", "slot3", "slot4", "bb", "nic", "bmc", "nicexp"};

int
fby3_common_get_gpio_shadow_array(const char **shadows, int num, uint8_t *mask) {
  int i;
  *mask = 0;
  for (i = 0; i < num; i++) {
    int ret;
    gpio_value_t value;
    gpio_desc_t *gpio = gpio_open_by_shadow(shadows[i]);
    if (!gpio) {
      return -1;
    }
    ret = gpio_get_value(gpio, &value);
    gpio_close(gpio);
    if (ret != 0) {
      return -1;
    }
    *mask |= (value == GPIO_VALUE_HIGH ? 1 : 0) << i;
  }
  return 0;
}

static int
set_gpio_value(const char *gpio_name, gpio_value_t val) {
  int ret = 0;
  gpio_desc_t *gdesc = NULL;
 
  gdesc = gpio_open_by_shadow(gpio_name);
  if ( gdesc == NULL ) {
    syslog(LOG_WARNING, "[%s] Cannot open %s", __func__, gpio_name);
    return -1;
  }

  ret = gpio_set_value(gdesc, val);
  gpio_close(gdesc);
  if ( ret < 0 ) {
    syslog(LOG_WARNING, "[%s] Cannot set %s to %d", __func__, gpio_name, val);
  }

  return ret;
}

static int
get_gpio_value(const char *gpio_name, uint8_t *status) {
  int ret = 0;
  gpio_desc_t *gdesc = NULL;
  gpio_value_t val;

  gdesc = gpio_open_by_shadow(gpio_name);
  if ( gdesc == NULL ) {
    syslog(LOG_WARNING, "[%s] Cannot open %s", __func__, gpio_name);
    return -1;
  }

  ret = gpio_get_value(gdesc, &val);
  gpio_close(gdesc);
  if ( ret < 0 ) {
    syslog(LOG_WARNING, "[%s] Cannot get the value of %s", __func__, gpio_name);
  }

  *status = (val == GPIO_VALUE_HIGH ? 1 : 0);
  return ret;
}

int
fby3_common_set_fru_i2c_isolated(uint8_t fru, uint8_t val) {
  return set_gpio_value(gpio_server_i2c_isolated[fru], (val==0)?GPIO_VALUE_LOW:GPIO_VALUE_HIGH);
}

int
fby3_common_get_fru_id(char *str, uint8_t *fru) {
  int fru_id = 0;
  bool found_id = false;

  for (fru_id = FRU_ALL; fru_id <= MAX_NUM_FRUS; fru_id++) {
    if ( strcmp(str, slot_list[fru_id]) == 0 ) {
      *fru = fru_id;
      found_id = true;
      break;
    }
  }

  if ( found_id == false ) {
    return -1;
  }

  return 0;
}

int
fby3_common_get_slot_id(char *str, uint8_t *fru) {
  int fru_id = 0;
  bool found_id = false;

  for (fru_id = FRU_ALL; fru_id <= FRU_SLOT4; fru_id++) {
    if ( strcmp(str, slot_list[fru_id]) == 0 ) {
      *fru = fru_id;
      found_id = true;
      break;
    }
  }

  if ( found_id == false ) {
    return -1;
  }

  return 0;
}

int 
fby3_common_check_slot_id(uint8_t fru) {
  switch (fru) {
    case FRU_SLOT1:
    case FRU_SLOT2:
    case FRU_SLOT3:
    case FRU_SLOT4:
      return 0;
    break;
  }

  return -1;
}

int
fby3_common_is_fru_prsnt(uint8_t fru, uint8_t *val) {
  int ret = 0;
  uint8_t bmc_location = 0;

  ret = fby3_common_get_bmc_location(&bmc_location);
  if ( ret < 0 ) {
    syslog(LOG_WARNING, "%s() Cannot get the location of BMC", __func__);
    return ret;
  }

  if ( bmc_location == BB_BMC || bmc_location == DVT_BB_BMC ) {
    ret = get_gpio_value(gpio_server_prsnt[fru], val);
    if ( ret < 0 ) {
      return ret;
    }

    //0: the fru isn't present
    //1: the fru is present
    *val = (*val == GPIO_VALUE_HIGH)?0:1;
  } else {
    //1: a server is always existed on class2
    *val = 1;
  }

  return ret; 
}

int
fby3_common_server_stby_pwr_sts(uint8_t fru, uint8_t *val) {
  int ret = 0;
  uint8_t bmc_location = 0;

  ret = fby3_common_get_bmc_location(&bmc_location);
  if ( ret < 0 ) {
    syslog(LOG_WARNING, "%s() Cannot get the location of BMC", __func__);
    return ret;
  }

  if ( bmc_location == BB_BMC || bmc_location == DVT_BB_BMC ) {
    ret = get_gpio_value(gpio_server_stby_pwr_sts[fru], val);
    if ( ret < 0 ) {
      return ret;
    }
  } else {
    //1: a server is always existed on class2 
    *val = 1;
  }

  return ret;
}

int
fby3_common_is_bic_ready(uint8_t fru, uint8_t *val) {
  int i2cfd = 0;
  int ret = 0;
  char path[32] = {0};
  uint8_t bmc_location = 0;
  uint8_t bus = 0;
  uint8_t tbuf[1] = {0x02};
  uint8_t rbuf[1] = {0};
  uint8_t tlen = 1;
  uint8_t rlen = 1;

  ret = fby3_common_get_bmc_location(&bmc_location);
  if ( ret < 0 ) {
    syslog(LOG_WARNING, "%s() Cannot get the location of BMC", __func__);
    goto error_exit;
  }

  //a bus starts from 4 
  ret = fby3_common_get_bus_id(fru) + 4;
  if ( ret < 0 ) {
    syslog(LOG_WARNING, "%s() Cannot get the bus with fru%d", __func__, fru);
    goto error_exit;
  }

  bus= (uint8_t)ret;
  snprintf(path, sizeof(path), I2CDEV, bus);

  i2cfd = open(path, O_RDWR);
  if ( i2cfd < 0 ) {
    syslog(LOG_WARNING, "%s() Failed to open %s", __func__, path);
    goto error_exit;
  }

  ret = i2c_rdwr_msg_transfer(i2cfd, (SB_CPLD_ADDR << 1), tbuf, tlen, rbuf, rlen);
  if ( ret < 0 ) {
    syslog(LOG_WARNING, "%s() Failed to do i2c_rdwr_msg_transfer, tlen=%d", __func__, tlen);
    goto error_exit;
  }

  *val = (rbuf[0] & 0x2) >> 1;
  
error_exit:
  if ( i2cfd > 0 ) {
    close(i2cfd);
  }

  return ret;
}

int
fby3_common_check_sled_mgmt_cbl_id(uint8_t slot_id, uint8_t *cbl_val, bool log_evnt, uint8_t bmc_location) {
  enum {
    SLOT1_CBL = 0x03,
    SLOT2_CBL = 0x02,
    SLOT3_CBL = 0x01,
    SLOT4_CBL = 0x00,
  };
  const uint8_t mapping_tbl[4] = {SLOT1_CBL, SLOT2_CBL, SLOT3_CBL, SLOT4_CBL};
  const char *gpio_mgmt_cbl_tbl[] = {"SLOT%d_ID1_DETECT_BMC_N", "SLOT%d_ID0_DETECT_BMC_N"};
  const int num_of_mgmt_pins = ARRAY_SIZE(gpio_mgmt_cbl_tbl);
  int i = 0;
  int ret = 0;
  char dev[32] = {0};
  uint8_t val = 0;
  uint8_t gpio_vals = 0;

  if ( bmc_location == DVT_BB_BMC ) {
    //read GPIO vals
    for ( i = 0; i < num_of_mgmt_pins; i++ ) {
      snprintf(dev, sizeof(dev), gpio_mgmt_cbl_tbl[i], slot_id);
      if ( get_gpio_value(dev, &val) < 0 ) {
        syslog(LOG_WARNING, "%s() Failed to read %s", __func__, dev);
      }
      gpio_vals |= (val << i);
    }
  } else {
    //NIC EXP
    //read GPIO from BB BIC
    //read 06h from SB CPLD
  }

  bool vals_match = (gpio_vals == mapping_tbl[slot_id-1]);
  if ( log_evnt == true ) {
    syslog(LOG_CRIT, "The cable of slot%d is %s", slot_id, (vals_match == false)?"mismatched":"matched");
  }

  if ( cbl_val != NULL ) *cbl_val = (vals_match == false)?STATUS_ABNORMAL:STATUS_PRSNT;
  return ret;
}

int
fby3_common_get_bus_id(uint8_t slot_id) {
  int bus_id = 0;
  switch(slot_id) {
    case FRU_SLOT1:
      bus_id = IPMB_SLOT1_I2C_BUS;
    break;
    case FRU_SLOT2:
      bus_id = IPMB_SLOT2_I2C_BUS;
    break;
    case FRU_SLOT3:
      bus_id = IPMB_SLOT3_I2C_BUS;
    break;
    case FRU_SLOT4:
      bus_id = IPMB_SLOT4_I2C_BUS;
    break;
    default:
      bus_id = -1;
    break;
  }
  return bus_id; 
}

int
fby3_common_get_bmc_location(uint8_t *id) {
  static bool is_cached = false;
  static uint8_t cached_id = 0;

  if ( is_cached == false ) {
    const char *shadows[] = {
      "BOARD_BMC_ID0_R",
      "BOARD_BMC_ID1_R",
      "BOARD_BMC_ID2_R",
      "BOARD_BMC_ID3_R",
    };

    if ( fby3_common_get_gpio_shadow_array(shadows, ARRAY_SIZE(shadows), &cached_id) ) {
      return -1;
    }

    is_cached = true;
  }

  *id = cached_id;

  return 0;
}

int
fby3_common_get_slot_type(uint8_t fru) {
  int ret = 0;

  ret = fby3_common_check_slot_id(fru);
  if ( ret < 0 ) {
    syslog(LOG_WARNING, "%s() Unknown fru: %d", __func__, fru);
    return -1;
  }

  return SERVER_TYPE_DL;
}

#define CRASHDUMP_BIN       "/usr/local/bin/autodump.sh"

int
fby3_common_crashdump(uint8_t fru, bool ierr, bool platform_reset) {
  int ret = 0;
  char cmd[128] = "\0";
  int cmd_len = sizeof(cmd);

  // Check if the crashdump script exist
  if (access(CRASHDUMP_BIN, F_OK) == -1) {
    syslog(LOG_CRIT, "%s() Try to run autodump for %d but %s is not existed", __func__, fru, CRASHDUMP_BIN);
    return ret;
  }

  if ( platform_reset == true ) {
    snprintf(cmd, cmd_len, "%s slot%d --second &", CRASHDUMP_BIN, fru);
  } else {
    snprintf(cmd, cmd_len, "%s slot%d &", CRASHDUMP_BIN, fru);
  }
  
  if ( system(cmd) != 0 ) {
    syslog(LOG_INFO, "%s() Crashdump for FRU: %d is failed to be generated.", __func__, fru);
  } else {
    syslog(LOG_INFO, "%s() Crashdump for FRU: %d is being generated.", __func__, fru);
  }
  return 0;
}

int
fby3_common_dev_id(char *str, uint8_t *dev) {

  if (!strcmp(str, "all")) {
    *dev = FRU_ALL;
  } else if (!strcmp(str, "1U-dev0")) {
    *dev = DEV_ID0_1OU;
  } else if (!strcmp(str, "1U-dev1")) {
    *dev = DEV_ID1_1OU;
  } else if (!strcmp(str, "1U-dev2")) {
    *dev = DEV_ID2_1OU;
  } else if (!strcmp(str, "1U-dev3")) {
    *dev = DEV_ID3_1OU;
  } else if (!strcmp(str, "2U-dev0")) {
    *dev = DEV_ID0_2OU;
  } else if (!strcmp(str, "2U-dev1")) {
    *dev = DEV_ID1_2OU;
  } else if (!strcmp(str, "2U-dev2")) {
    *dev = DEV_ID2_2OU;
  } else if (!strcmp(str, "2U-dev3")) {
    *dev = DEV_ID3_2OU;
  } else if (!strcmp(str, "2U-dev4")) {
    *dev = DEV_ID4_2OU;
  } else if (!strcmp(str, "2U-dev5")) {
    *dev = DEV_ID5_2OU;
  } else if (!strcmp(str, "1U")) {
    *dev = BOARD_1OU;
  } else if (!strcmp(str, "2U")) {
    *dev = BOARD_2OU;
  } else {
#ifdef DEBUG
    syslog(LOG_WARNING, "fby3_common_dev_id: Wrong fru id");
#endif
    return -1;
  }

  return 0;
}

int
fby3_common_dev_name(uint8_t dev, char *str) {

  if (dev == FRU_ALL) {
    strcpy(str, "all");
  } else if (dev == DEV_ID0_1OU) {
    strcpy(str, "1U-dev0");
  } else if (dev == DEV_ID1_1OU) {
    strcpy(str, "1U-dev1");
  } else if (dev == DEV_ID2_1OU) {
    strcpy(str, "1U-dev2");
  } else if (dev == DEV_ID3_1OU) {
    strcpy(str, "1U-dev3");
  } else if (dev == DEV_ID0_2OU) {
    strcpy(str, "2U-dev0");
  } else if (dev == DEV_ID1_2OU) {
    strcpy(str, "2U-dev1");
  } else if (dev == DEV_ID2_2OU) {
    strcpy(str, "2U-dev2");
  } else if (dev == DEV_ID3_2OU) {
    strcpy(str, "2U-dev3");
  } else if (dev == DEV_ID4_2OU) {
    strcpy(str, "2U-dev4");
  } else if (dev == DEV_ID5_2OU) {
    strcpy(str, "2U-dev5");
  } else if (dev == BOARD_1OU) {
    strcpy(str, "1U");
  } else if (dev == BOARD_2OU) {
    strcpy(str, "2U");
  } else {
#ifdef DEBUG
    syslog(LOG_WARNING, "fby3_common_dev_id: Wrong fru id");
#endif
    return -1;
  }
  return 0;
}
