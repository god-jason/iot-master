let _redirectingToLogin = false

export const navigateToLogin = () => {
  if (_redirectingToLogin) return
  _redirectingToLogin = true
  try {
    uni.reLaunch({
      url: '/pages/login/index',
      complete: () => {
        setTimeout(() => { _redirectingToLogin = false }, 800)
      }
    })
  } catch (e) {
    _redirectingToLogin = false
  }
}

export const resetNavigationLock = () => {
  _redirectingToLogin = false
}

export const isRedirectingToLogin = () => _redirectingToLogin
