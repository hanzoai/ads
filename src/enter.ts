/** Sign in with a popup so the page stays put; a blocked popup falls back to a redirect. */
export interface Door {
  login: () => unknown
  loginPopup: () => Promise<unknown>
}

export async function enter(door: Door): Promise<boolean> {
  try {
    await door.loginPopup()
    return true
  } catch {
    void door.login()
    return false
  }
}
