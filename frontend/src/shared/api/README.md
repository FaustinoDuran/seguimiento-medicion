Cliente HTTP hacia la API Go. Sin reglas de negocio.

Importar `apiRequest` desde `./client`. Recibe una ruta de la API sin `/api`,
opciones de `fetch` (incluido `signal`) y un `body` opcional serializado a JSON.

```ts
const health = await apiRequest<{ status: string }>('/health', { signal })
const project = await apiRequest<Project>('/projects', {
  method: 'POST',
  body: { name, start_date, end_date },
})
```

Devuelve JSON; una respuesta vacía devuelve `undefined` (usar `<void>` para
operaciones sin cuerpo). Los errores HTTP lanzan `ApiError` con `status`,
`body` y el mensaje `error` enviado por Go cuando existe. Los errores de red
y cancelación se propagan desde `fetch`. El tipo genérico describe el contrato;
no valida los datos en tiempo de ejecución ni transforma reglas del dominio.
