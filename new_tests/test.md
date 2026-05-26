Tabla de casos de prueba de gopdfsuit
TC ID
Tier
Escenario
Dato de entrada
Precondiciones de ejecución
Salida esperada
Pass/Fail
01
go-be
Verificar que el PDF es rellenable a través de /api/v1/redact/capabilities.
us_patient_healthcare_form_compressed.pdf
El endpoint está aislado y el servidor funciona.
Objeto JSON con is_fillable: true y código HTTP 200.
Pass
02
fe-react
Frontend intenta llenar el AcroForm usando mocks.
Archivos File en React para PDF y XFDF.
El backend API (/api/v1/fill) está mockeado con una respuesta exitosa (200).
Se procesa el blob mockeado y se invoca la descarga (click).
Pass
03
fe-react
Frontend falla al llenar el AcroForm por error del servidor.
Archivos File en React para PDF y XFDF.
El backend API está mockeado para devolver un error HTTP 500.
La función maneja el error mostrando una alerta al usuario sin romper el frontend.
Fail
04
go-be
El backend rellena el formulario de manera independiente en /api/v1/fill.
Form-data con PDF y XFDF sintético válidos.
Las dependencias del endpoint están aisladas. Es decir, probar el comportamiento del endpoint sin depender de un entorno externo ni de piezas que no controlamos en el test. No se depende de otros servicios para que /api/v1/fill responda.
Archivo PDF con los campos rellenados y código HTTP 200.
Pass
05
fe-react
Frontend solicita la generación de portada.
Configuración JSON para template.
El backend API (/api/v1/generate/template-pdf) está mockeado de forma exitosa.
La promesa resuelve y el frontend genera o descarga el PDF simulado.
Pass
06
go-be
Generación backend de portada con payload vacío en /generate/template-pdf.
Cuerpo de la petición vacío o JSON incompleto.
Servidor aislado.
Error al parsear el JSON y código HTTP 400.
Fail
07
fe-react
Frontend une la portada y el PDF lleno.
Objeto FormData con los dos PDFs mockeados en estado.
El backend API (/api/v1/merge) está mockeado retornando Blob validado.
Se invoca exitosamente la descarga del bundle consolidado.
Pass
08
go-be
Fusión exitosa de documentos en el backend /api/v1/merge.
Dos archivos PDF pasados por form-data.
Ambos archivos aislados están libres de corrupción.
Documento application/pdf unificado y código HTTP 200.
Pass
09
go-be
Validación de estructura del bundle re-encriptado en /api/v1/redact/page-info.
El archivo PDF bundle re-encriptado aisladamente.
El bundle fue procesado y encriptado sin alterar el parseo interno.
Objeto JSON validando las páginas y código HTTP 200.
Pass
10
go-be
Intento de validar un bundle cuya encriptación lo corrompió.
Bundle PDF ilegible (bytes rotos).
El proceso de encriptación intencionalmente estropeó la estructura.
Error de validación del PDF y código HTTP 400/500.
Fail


Ejecutar los tests:
Frontend
cd /home/fnovoas/gopdfsuit/frontend && npm run test:isolated

Backend
cd /home/fnovoas/gopdfsuit && go test -v ./new_tests/backend/... -count=1

Obtenemos:

-----------------------
TC ID
Nivel / fronteras
Escenario
Dato de entrada
Pasos
Salida esperada
Pass/Fail
01
Backend ↔ Frontend (contrato HTTP multipart/form-data sobre POST /api/v1/merge)
Verificar que cuando un usuario sube varios PDFs en la página de merge y los reordena desde la interfaz, el PDF combinado que descarga tiene las páginas en el mismo orden que él eligió.
3 PDFs de 1 página cada uno, identificables por una marca de texto distinta en cada uno: el primero contiene la palabra "MARK_A", el segundo "MARK_B" y el tercero "MARK_C". El usuario los sube en ese orden (A, B, C) y luego los reordena en el frontend para que queden en orden C, A, B. 
1) Abrir la vista /gopdfsuit/#/merge.
2) Subir los 3 PDFs (A, B, C).
3) Reordenarlos en pantalla hasta que queden en orden C, A, B.
4) Presionar "Merge PDFs".
5) El frontend envía la petición al backend.
6) Descargar el PDF resultante.
7) Abrir el PDF descargado y revisar el contenido de cada página.
Un único PDF con 3 páginas: la primera página contiene "MARK_C", la segunda "MARK_A" y la tercera "MARK_B" - en ese orden exacto, coincidiendo con el orden que el usuario definió en el frontend. 
Pass
02
Backend ↔ Backend (contrato AcroForm entre POST /api/v1/generate/template-pdf y POST /api/v1/fill)
Generar un PDF con form fields desde un template JSON y llenarlo vía XFDF a través de los endpoints HTTP reales, sin invocar las funciones internas directamente.
Plantilla JSON sampledata/acroform/us_patient_healthcare_form_og.json (contiene el campo first_name) + XFDF construido en memoria con
<field name="first_name">
<value>TestNombre</value>
</field>.


 1) POST /api/v1/generate/template-pdf con el JSON como body → recibir bytes del PDF. 
2) Verificar que la respuesta es application/pdf y empieza con %PDF. 
3) POST /api/v1/fill como multipart/form-data con pdf = bytes generados y xfdf = XFDF construido. 
4) Leer los bytes de la respuesta. 5) Buscar TestNombre en los bytes crudos del PDF. 
6) Buscar /V (TestNombre) para confirmar que el valor cayó dentro de un campo AcroForm.


Status 200 en ambas peticiones. El PDF resultante contiene la subcadena TestNombre y específicamente /V TestNombre), evidenciando que el valor llegó al /V del campo y no a otra parte del documento.


Pass 
03
Frontend ↔ Backend (módulo internal/pdf)

flujo E2E sobre POST /api/v1/merge
Verificar que al intentar hacer merge entre dos pdf corruptos desde la interfaz web, el backend rechaza la operación y el frontend muestra correctamente el manejo de error sin generar preview ni descarga.
2 archivos corrupted.pdf generados en memoria con bytes que no representan un pdf válido
1) Abrir la vista /gopdfsuit/#/merge.
 2) Subir corrupted1.pdf y corrupted2.pdf. 
3) Verificar que ambos archivos aparecen en la lista de la UI.
4) Registrar page.on('dialog', ...) para capturar el alert. 
5) Presionar "Merge PDFs". 
6) Esperar la respuesta real del backend sobre /api/v1/merge.
7) Verificar el estado de la interfaz después del error.
El backend responde con status  500 y Content-Type: application/json. El frontend muestra un alert que contiene "Error merging PDFs". No se genera iframe [title="Merged PDF"], no se dispara ninguna descarga y el botón "Merge PDFs" vuelve al estado habilitado después del fallo
Fail
04 
End to end (E2E)
Se está verificando el flujo end-to-end esperado del backend la prueba comprueba que el servidor levanta correctamente, que la ruta raíz redirige a /gopdfsuit, que la SPA responde en esa ruta y que el endpoint /api/v1/generate/template-pdf genera un PDF válido a partir de un JSON real del repo. 
que el backend realmente arranca por HTTP local;
que las rutas públicas principales están conectadas;
que el handler de generación de PDF procesa un payload real;
que la salida es un PDF, no solo una respuesta HTTP exitosa.
Archivo JSON: sampledata/editor/financial_digitalsignature.json
1) Se instancia un servidor HTTP real con gin.New() y handlers.RegisterRoutes(router), sin usar httptest.NewServer, para que la ejecución pase por el stack de red de net/http y no por un atajo de pruebas, en esta misma parte ell servidor se enlaza a 127.0.0.1:0, dejando que el sistema asigne un puerto libre.
2) Se ejecuta con server.Serve(listener).
3) Se implementa un polling de readiness contra GET / hasta que el servidor responde, asegurando que el test no avance antes de que el backend esté aceptando conexiones.
4) Se invoca GET / con redirecciones deshabilitadas para verificar que el handler raíz responde con 301 Moved Permanently y cabecera Location: /gopdfsuit.
5) Se invoca GET /gopdfsuit para validar que la SPA compilada se sirve correctamente desde docs/index.html y que la respuesta devuelve 200 OK con contenido HTML.
6) Se carga desde disco un payload JSON real del repositorio y se envía por POST /api/v1/generate/template-pdf con Content-Type: application/json.
Se valida la respuesta a nivel de contrato HTTP: 200 OK, Content-Type: application/pdf y Content-Disposition con nombre de archivo generated.pdf.
7) Se lee el body completo de la respuesta y se comprueba que el binario no esté vacío, supere un umbral mínimo de tamaño y comience con la firma %PDF-, confirmando que el servidor generó un PDF válido.
GET / responde 301 y Location: /gopdfsuit.
GET /gopdfsuit responde 200 con HTML.
POST /api/v1/generate/template-pdf responde 200.
Content-Type: application/pdf.
Content-Disposition incluye generated.pdf.
El body empieza con %PDF- y no está vacío.
Que se genere un archivo .pdf válido. 
Pass
05
Frontend → Google OAuth → Backend

Test E2E que integra un flujo compuesto de tres componentes.
Verificar el login manejado con google OAuth desde el frontend, verificando la integración con el sistema OAuth externo de google, posteriormente verificar en los endpoints del backend, se crea un double /api/v1/test/auth que usará la función que se encarga de verificar los headers de autenticación.
- Variable de entorno GOOGLE_CLIENT_ID
- Credenciales de usuario para flujo real, se utilizan credenciales para las pruebas.
1) Abrir la vista de login
2) Simular login usando las credenciales de prueba
3) Obtener respuesta de Google OAuth y obtener los tokens de autenticación
4) La función de test del front llamará a /api/v1/test/auth del backend utilizando las funciones actuales para poner los tokens en el header
5) Obtener respuesta json del backend sobre verificación de autenticación.
De Google OAuth:
- Objeto JSON con los tokens de autenticación

De Backend:
- Objeto JSON que confirma la verificación de la validez de los tokens con Google OAuth.
Pass

--------------------------------
