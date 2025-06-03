/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export const redirectToLogin = (shouldRecordUrl = true) => {
  if (shouldRecordUrl) {
    const currentUrl = window.location.href;
    sessionStorage.setItem('urlBeforeLogin', currentUrl);
  }
  window.location.href = '/#/login';
};