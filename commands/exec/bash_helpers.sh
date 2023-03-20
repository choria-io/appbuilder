# Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
#
# SPDX-License-Identifier: Apache-2.0

AB_SAY_PREFIX=">>>"
AB_ERROR_PREFIX="!!!"

function ab_prefix() {
  local p="$1"

  case "$AB_HELPER_TIME_STAMP" in
    T | t | 1)
      p="[$(date '+%T')] ${p}"
      ;;
    D | d | 2)
      p="[$(date '+%F %T')] ${p}"
      ;;
  esac

  echo "${p}"
}

function ab_say() {
  local p

  p=$(ab_prefix "${AB_SAY_PREFIX}")
  echo "${p} $*"
}

function ab_announce() {
  local p

  p=$(ab_prefix "${AB_SAY_PREFIX}")

  echo "${p}"
  echo "${p} $*"
  echo "${p}"
}

function ab_error() {
  local p

  p=$(ab_prefix "${AB_ERROR_PREFIX}")

  echo "${p}"
  echo "${p} $*"
  echo "${p}"
}

function ab_panic() {
  error "$*"
  exit 1
}
