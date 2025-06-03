export const redirectToLogin = (shouldRecordUrl = true) => {
  if (shouldRecordUrl) {
    const currentUrl = window.location.href;
    sessionStorage.setItem('urlBeforeLogin', currentUrl);
  }
  window.location.href = '/#/login';
};