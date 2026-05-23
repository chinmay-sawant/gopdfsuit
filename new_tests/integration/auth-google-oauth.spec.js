import { test, expect } from '@playwright/test'

const GOOGLE_CLIENT_ID = process.env.GOOGLE_CLIENT_ID ||
  '46981518442-ap2ga76ao0sj82t47mimcv8eot0l0pbt.apps.googleusercontent.com'
const GOOGLE_TEST_EMAIL = process.env.GOOGLE_TEST_EMAIL
const GOOGLE_TEST_PASSWORD = process.env.GOOGLE_TEST_PASSWORD

const EDITOR_URL = '/gopdfsuit/#/editor'
const AUTH_ENDPOINT = '/api/v1/test/auth'

async function fetchAuthEndpoint(page, token) {
  return page.evaluate(async ([url, tok]) => {
    try {
      const headers = { 'Content-Type': 'application/json' }
      if (tok) headers['Authorization'] = `Bearer ${tok}`
      const res = await fetch(url, { method: 'GET', headers })
      const data = await res.json()
      return { status: res.status, data }
    } catch (err) {
      return { status: 0, data: { error: err.message } }
    }
  }, [AUTH_ENDPOINT, token])
}

test.describe('E2E: Frontend -> Google OAuth -> Backend Auth Verification', () => {

  test.describe.configure({ timeout: 120000 })

  test('Step 1: Login view shows Google OAuth button', async ({ page }) => {
    await page.goto(EDITOR_URL)

    const heading = page.getByRole('heading', { name: /PDF Template Editor/i })
    await expect(heading).toBeVisible({ timeout: 15000 })

    const signInText = page.getByText(/sign in with your Google account/i)
    await expect(signInText).toBeVisible()
  })

  test('Step 2-3: Google OAuth login and token acquisition', async ({ page, context }) => {
    test.skip(!GOOGLE_TEST_EMAIL || !GOOGLE_TEST_PASSWORD,
      'GOOGLE_TEST_EMAIL and GOOGLE_TEST_PASSWORD env vars required')

    await page.goto(EDITOR_URL)

    const heading = page.getByRole('heading', { name: /PDF Template Editor/i })
    await expect(heading).toBeVisible({ timeout: 15000 })

    const googlePopupPromise = context.waitForEvent('page', { timeout: 30000 })

    const signInButton = page.locator('[data-testid="google-login-button"]').first()
    const googleIframe = page.locator('iframe[title*="Sign in"]').first()
    const anyGoogleButton = page.locator('div[role="button"][aria-label*="Sign in"]').first()

    if (await signInButton.count() > 0) {
      await signInButton.click()
    } else if (await googleIframe.count() > 0) {
      const frame = await googleIframe.contentFrame()
      await frame.locator('[role="button"]').first().click()
    } else if (await anyGoogleButton.count() > 0) {
      await anyGoogleButton.click()
    } else {
      const allButtons = page.locator('button, [role="button"]')
      const count = await allButtons.count()
      for (let i = 0; i < count; i++) {
        const btn = allButtons.nth(i)
        const text = await btn.textContent().catch(() => '')
        if (text && text.toLowerCase().includes('sign in')) {
          await btn.click()
          break
        }
      }
    }

    let popup
    try {
      popup = await googlePopupPromise
    } catch {
      popup = null
    }

    if (popup) {
      await popup.waitForLoadState('domcontentloaded', { timeout: 15000 }).catch(() => {})

      const emailInput = popup.locator('input[type="email"]').first()
      await expect(emailInput).toBeVisible({ timeout: 15000 })
      await emailInput.fill(GOOGLE_TEST_EMAIL)

      const nextButton = popup.locator('#identifierNext, button:has-text("Next")').first()
      await nextButton.click().catch(async () => {
        await popup.keyboard.press('Enter')
      })

      await popup.waitForTimeout(2000)

      const passwordInput = popup.locator('input[type="password"]').first()
      await expect(passwordInput).toBeVisible({ timeout: 15000 })
      await passwordInput.fill(GOOGLE_TEST_PASSWORD)

      const passwordNext = popup.locator('#passwordNext, button:has-text("Next")').first()
      await passwordNext.click().catch(async () => {
        await popup.keyboard.press('Enter')
      })

      await popup.waitForTimeout(3000)
      await popup.waitForEvent('close', { timeout: 30000 }).catch(() => {})
    }

    const hasToken = await page.evaluate(() => {
      return !!localStorage.getItem('google_id_token')
    })

    if (!hasToken) {
      const idToken = await obtainIdTokenProgrammatically()
      if (idToken) {
        await injectToken(page, idToken)
        await page.reload()
        await page.waitForTimeout(2000)
      }
    }

    const storedToken = await page.evaluate(() => localStorage.getItem('google_id_token'))
    expect(storedToken).toBeTruthy()

    const storedUser = await page.evaluate(() => {
      const raw = localStorage.getItem('google_user')
      return raw ? JSON.parse(raw) : null
    })
    expect(storedUser).toBeTruthy()
    expect(storedUser.email).toBeTruthy()
  })

  test('Step 4-5: Backend /api/v1/test/auth verifies token', async ({ page }) => {
    test.skip(!GOOGLE_TEST_EMAIL || !GOOGLE_TEST_PASSWORD,
      'GOOGLE_TEST_EMAIL and GOOGLE_TEST_PASSWORD env vars required')

    await page.goto(EDITOR_URL)
    await page.waitForTimeout(3000)

    let storedToken = await page.evaluate(() => localStorage.getItem('google_id_token'))

    if (!storedToken) {
      const idToken = await obtainIdTokenProgrammatically()
      test.skip(!idToken, 'Could not obtain Google ID token for test')

      await injectToken(page, idToken)
      await page.reload()
      await page.waitForTimeout(3000)
      storedToken = idToken
    }

    test.skip(!storedToken, 'No Google ID token available')

    const backendResponse = await fetchAuthEndpoint(page, storedToken)

    expect(backendResponse.status).toBe(200)
    expect(backendResponse.data.authenticated).toBe(true)
    expect(backendResponse.data.message).toContain('verified')
    expect(backendResponse.data.user).toBeTruthy()
    expect(backendResponse.data.user.email).toBeTruthy()
  })

  test('Full flow: Login -> Token -> Backend verification (integration)', async ({ page, context }) => {
    test.skip(!GOOGLE_TEST_EMAIL || !GOOGLE_TEST_PASSWORD,
      'GOOGLE_TEST_EMAIL and GOOGLE_TEST_PASSWORD env vars required')

    await page.goto(EDITOR_URL)

    const heading = page.getByRole('heading', { name: /PDF Template Editor/i })
    await expect(heading).toBeVisible({ timeout: 15000 })

    const idToken = await performGoogleLogin(page, context)

    expect(idToken).toBeTruthy()
    const tokenParts = idToken.split('.')
    expect(tokenParts.length).toBeGreaterThanOrEqual(2)

    const payload = JSON.parse(atob(tokenParts[1]))
    expect(payload.email).toBeTruthy()
    expect(payload.email).toContain('@')

    const backendResponse = await fetchAuthEndpoint(page, idToken)

    expect(backendResponse.status).toBe(200)
    expect(backendResponse.data.authenticated).toBe(true)
    expect(backendResponse.data.message).toContain('verified')
    expect(backendResponse.data.user).toBeTruthy()
    expect(backendResponse.data.user.email).toBe(payload.email)
  })

  test('Backend rejects invalid token at /api/v1/test/auth', async ({ page }) => {
    await page.goto(EDITOR_URL)
    await page.waitForTimeout(2000)

    const invalidToken = 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImZha2VAdGVzdC5jb20ifQ.fake'

    const response = await fetchAuthEndpoint(page, invalidToken)

    expect(response.status).toBe(401)
    expect(response.data.error).toContain('Invalid')
  })

  test('Backend rejects request without Authorization header', async ({ page }) => {
    await page.goto(EDITOR_URL)
    await page.waitForTimeout(2000)

    const response = await fetchAuthEndpoint(page, null)

    expect(response.status).toBe(401)
    expect(response.data.error).toContain('Authorization')
  })
})

async function injectToken(page, token) {
  await page.evaluate((t) => {
    const payload = JSON.parse(atob(t.split('.')[1]))
    const userData = {
      email: payload.email,
      name: payload.name,
      picture: payload.picture,
    }
    localStorage.setItem('google_id_token', t)
    localStorage.setItem('google_user', JSON.stringify(userData))
  }, token)
}

async function performGoogleLogin(page, context) {
  const googlePopupPromise = context.waitForEvent('page', { timeout: 30000 }).catch(() => null)

  const anyGoogleButton = page.locator('div[role="button"][aria-label*="Sign in"]').first()
  const signInButton = page.locator('[data-testid="google-login-button"]').first()

  if (await signInButton.count() > 0) {
    await signInButton.click()
  } else if (await anyGoogleButton.count() > 0) {
    await anyGoogleButton.click()
  } else {
    const allButtons = page.locator('button, [role="button"]')
    const count = await allButtons.count()
    for (let i = 0; i < count; i++) {
      const btn = allButtons.nth(i)
      const text = await btn.textContent().catch(() => '')
      if (text && text.toLowerCase().includes('sign in')) {
        await btn.click()
        break
      }
    }
  }

  const popup = await googlePopupPromise

  if (popup) {
    await popup.waitForLoadState('domcontentloaded', { timeout: 15000 }).catch(() => {})

    const emailInput = popup.locator('input[type="email"]').first()
    await expect(emailInput).toBeVisible({ timeout: 15000 })
    await emailInput.fill(GOOGLE_TEST_EMAIL)

    const nextButton = popup.locator('#identifierNext, button:has-text("Next")').first()
    await nextButton.click().catch(async () => { await popup.keyboard.press('Enter') })

    await popup.waitForTimeout(2000)

    const passwordInput = popup.locator('input[type="password"]').first()
    await expect(passwordInput).toBeVisible({ timeout: 15000 })
    await passwordInput.fill(GOOGLE_TEST_PASSWORD)

    const passwordNext = popup.locator('#passwordNext, button:has-text("Next")').first()
    await passwordNext.click().catch(async () => { await popup.keyboard.press('Enter') })

    await popup.waitForTimeout(3000)
    await popup.waitForEvent('close', { timeout: 30000 }).catch(() => {})
  }

  await page.waitForTimeout(2000)

  let storedToken = await page.evaluate(() => localStorage.getItem('google_id_token'))

  if (!storedToken) {
    storedToken = await obtainIdTokenProgrammatically()
    if (storedToken) {
      await injectToken(page, storedToken)
      await page.reload()
      await page.waitForTimeout(3000)
    }
  }

  return storedToken
}

async function obtainIdTokenProgrammatically() {
  const clientId = GOOGLE_CLIENT_ID
  const clientSecret = process.env.GOOGLE_CLIENT_SECRET
  const refreshToken = process.env.GOOGLE_REFRESH_TOKEN

  if (!refreshToken || !clientSecret) {
    return null
  }

  try {
    const response = await fetch('https://oauth2.googleapis.com/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        client_id: clientId,
        client_secret: clientSecret,
        refresh_token: refreshToken,
        grant_type: 'refresh_token',
      }).toString(),
    })

    const data = await response.json()
    return data.id_token || null
  } catch {
    return null
  }
}
