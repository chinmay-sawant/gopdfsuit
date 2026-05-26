# Estado de los tests tras introducir `auth-ms`

Este documento mapea los **10 unitarios + 5 de integración** definidos en
`new_tests/test.md` a su ubicación actual en el repo.

Reglas seguidas al armar este mapa:

- No se inventaron tests nuevos. Si un TC del doc no tiene spec dedicado,
  aparece como **No implementado**.
- El único TC que cambió por el reemplazo de Google es el **TC 05 de
  integración**. El contrato del test es el mismo (frontend obtiene token →
  backend lo verifica); cambia el proveedor (Google OAuth → auth-ms).
- La suite del microservicio nuevo (`auth-ms/auth_test.go`) **no forma parte
  de los 15 del doc**. Es un sanity check del propio servicio; se documenta
  aparte.

Comandos:

```
make test-unit          # los 10 unitarios + sanity de auth-ms
make test-integration   # los 5 de integración
make test               # ambos
```

---

## 10 unitarios

| TC | Tier      | Escenario                                                                      | Archivo                                       | Función / case                                            | Estado |
|----|-----------|--------------------------------------------------------------------------------|-----------------------------------------------|-----------------------------------------------------------|--------|
| 01 | go-be     | PDF rellenable en `POST /api/v1/redact/capabilities`                           | `new_tests/backend/flow_test.go`              | `TestFlowCases/TC 01 (Pass)`                              | ✅ Cumple |
| 02 | fe-react  | Frontend intenta llenar el AcroForm usando mocks                               | `new_tests/frontend/Filler.test.jsx`          | `TC 02 (Pass)`                                            | ✅ Cumple |
| 03 | fe-react  | Frontend falla al llenar el AcroForm por error del servidor                    | `new_tests/frontend/Filler.test.jsx`          | `TC 03 (Fail)`                                            | ✅ Cumple |
| 04 | go-be     | Backend rellena el formulario en `POST /api/v1/fill`                           | `new_tests/backend/flow_test.go`              | `TestFlowCases/TC 04 (Pass)`                              | ✅ Cumple |
| 05 | fe-react  | Frontend solicita la generación de portada                                     | `new_tests/frontend/Viewer.test.jsx`          | `TC 05 (Pass)`                                            | ✅ Cumple |
| 06 | go-be     | Generación con payload vacío en `POST /generate/template-pdf`                  | `new_tests/backend/flow_test.go`              | `TestFlowCases/TC 06 (Fail)`                              | ✅ Cumple |
| 07 | fe-react  | Frontend une la portada y el PDF lleno                                         | `new_tests/frontend/Merge.test.jsx`           | `TC 07 (Pass)`                                            | ✅ Cumple |
| 08 | go-be     | Fusión en `POST /api/v1/merge`                                                 | `new_tests/backend/flow_test.go`              | `TestFlowCases/TC 08 (Pass)`                              | ✅ Cumple |
| 09 | go-be     | Validación bundle re-encriptado en `POST /api/v1/redact/page-info`             | `new_tests/backend/flow_test.go`              | `TestFlowCases/TC 09 (Pass)`                              | ✅ Cumple |
| 10 | go-be     | Bundle cuya encriptación lo corrompió en `POST /api/v1/redact/page-info`       | `new_tests/backend/flow_test.go`              | `TestFlowCases/TC 10 (Fail)`                              | ✅ Cumple |

---

## 5 de integración

| TC | Nivel / fronteras                                                                                       | Herramienta             | Archivo                                          | Función / spec                                          | Estado |
|----|---------------------------------------------------------------------------------------------------------|-------------------------|--------------------------------------------------|---------------------------------------------------------|--------|
| 01 | Backend ↔ Frontend (multipart sobre `POST /api/v1/merge`) — reordenamiento del usuario                  | Go HTTP (httptest)      | `test/integration_test.go`                       | `TestIntegrationSuite/TestMergePreservesUserOrder`      | ✅ Cumple (traído de `fabio/tc01-merge-order`) |
| 02 | Backend ↔ Backend (contrato AcroForm entre `POST /api/v1/generate/template-pdf` y `POST /api/v1/fill`)  | Go HTTP (encadenado)    | `test/integration_test.go`                       | `TestIntegrationSuite/TestGenerateTemplatePDFThenFillWithXFDF` | ✅ Cumple (traído de `cat`) |
| 03 | Frontend ↔ Backend (flujo E2E sobre `POST /api/v1/merge`, camino fail con PDFs corruptos)               | Playwright (browser)    | `new_tests/integration/merge.fail.spec.js`       | `rechaza dos PDFs corruptos con 500`                    | ✅ Cumple |
| 04 | End-to-end del backend (servidor real + redirect + SPA + `POST /api/v1/generate/template-pdf`)          | Go HTTP (server real)   | `new_tests/backend/e2e_test.go`                  | `TestE2ETemplatePDFFlow`                                | ✅ Cumple |
| 05 | Frontend → **auth-ms** → Backend (antes Google OAuth)                                                   | Playwright (browser)    | `new_tests/integration/auth.spec.js`             | `TC 05: …login y verificación`                          | ✅ Cumple (proveedor reemplazado, contrato idéntico) |

### Detalle TC 05 — qué cambió y qué no

El doc original dice *"Frontend → Google OAuth → Backend"*. El nuevo spec
ejecuta **los mismos 5 pasos** descritos en el doc:

1. Abrir la vista de login.
2. Simular login usando las credenciales de prueba.
3. Obtener respuesta del servicio de auth y obtener los tokens.
4. El frontend llama a `/api/v1/test/auth` con los tokens en el header.
5. Recibir respuesta JSON del backend confirmando la validez del token.

Diferencia única: el "servicio de auth" ya no es Google sino nuestro `auth-ms`.
Beneficio: el test corre real en cada corrida (no necesita skips porque Google
bloquea automatización).

Stack que se levanta durante el test (lo configura
`new_tests/integration/playwright.cloudrun.config.js`):

- `auth-ms` :9090 — SQLite in-memory, JWT HS256 con secreto compartido.
- backend :8080 — `K_SERVICE=e2e-test` + mismo `AUTH_JWT_SECRET`.
- frontend :3000 — `VITE_IS_CLOUD_RUN=true`, `VITE_AUTH_URL=http://localhost:9090`.

---

## Sanity checks que no forman parte de los 15

Pertenecen al microservicio nuevo y se incluyen aparte:

| Archivo                                       | Qué cubre                                                            |
|-----------------------------------------------|----------------------------------------------------------------------|
| `auth-ms/auth_test.go`                        | Suite unitaria de `auth-ms` (register / login / verify / firma).     |
| `new_tests/backend/test_auth_test.go`         | Sanity del middleware `AuthMiddleware` (con y sin `K_SERVICE`).      |

Estos no aparecen en `test.md`; existen porque el ms es código nuevo
introducido por el reemplazo de Google OAuth.
