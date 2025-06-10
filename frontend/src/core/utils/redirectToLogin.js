/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export const redirectToLogin = (shouldRecordUrl = true) => {
  const loginRoute = '/#/login'

  if (shouldRecordUrl) {
    const pathWithQueryAndHash =
      window.location.pathname +
      window.location.search +
      window.location.hash;

    if (pathWithQueryAndHash !== loginRoute) {
      sessionStorage.setItem('urlBeforeLogin', pathWithQueryAndHash);
    }
  }
  window.location.href = '/#/login';
};