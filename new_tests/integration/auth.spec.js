import { test, expect } from '@playwright/test'

// TC 05 (integración): Frontend -> auth-ms -> Backend.
// Test E2E que integra un flujo compuesto de tres componentes.
//
// Original: "Frontend -> Google OAuth -> Backend". El proveedor cambió a
// nuestro auth-ms; el contrato del test es el mismo: el frontend obtiene un
// token desde el servicio de auth y el backend lo verifica vía
// /api/v1/test/auth, devolviendo JSON de verificación.
//
// Pasos (idénticos al test.md):
//   1) Abrir la vista de login.
//   2) Simular login usando las credenciales de prueba.
//   3) Obtener respuesta del servicio de auth y obtener los tokens.
//   4) El frontend llama a /api/v1/test/auth poniendo los tokens en el header.
//   5) Recibir respuesta JSON del backend confirmando la validez del token.

const EDITOR_URL = '/gopdfsuit/#/editor'
const AUTH_ENDPOINT = '/api/v1/test/auth'

test('TC 05: Frontend -> auth-ms -> Backend (login y verificación)', async ({ page }) => {
  // Email único por corrida — la DB del auth-ms es in-memory, pero igual
  // evitamos colisiones si el test corre dos veces sin reiniciar.
  const email = `tc05_${Date.now()}@example.com`
  const password = 'supersecret123'

  // 1) Abrir la vista de login.
  await page.goto(EDITOR_URL)
  await expect(page.getByRole('heading', { name: /PDF Template Editor/i })).toBeVisible()
  await expect(page.getByTestId('auth-form')).toBeVisible()

  // 2) Simular login usando las credenciales de prueba.
  //    Como auth-ms no permite Google OAuth, primero registramos al usuario
  //    (equivalente a "tener credenciales válidas en el proveedor") y luego
  //    el form lo deja autenticado con un token real.
  await page.getByTestId('auth-email').fill(email)
  await page.getByTestId('auth-password').fill(password)
  await page.getByTestId('auth-register').click()

  // 3) Obtener respuesta del servicio de auth y obtener los tokens.
  await expect(page.getByTestId('auth-form')).toHaveCount(0, { timeout: 15000 })
  const token = await page.evaluate(() => localStorage.getItem('auth_token'))
  expect(token).toBeTruthy()

  // 4) El frontend llama a /api/v1/test/auth poniendo los tokens en el header.
  // 5) Recibir respuesta JSON del backend confirmando la validez del token.
  const backendResponse = await page.evaluate(async ([url, tok]) => {
    const res = await fetch(url, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${tok}` },
    })
    return { status: res.status, data: await res.json() }
  }, [AUTH_ENDPOINT, token])

  expect(backendResponse.status).toBe(200)
  expect(backendResponse.data.authenticated).toBe(true)
  expect(backendResponse.data.user.email).toBe(email)
})
