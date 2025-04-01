
# 🚗 Nexu Challenge

Este proyecto es una solución al reto técnico de Nexu, donde se desarrolla un servicio backend para gestionar información de marcas y modelos de vehículos. El objetivo es exponer un conjunto de endpoints RESTful para consulta y administración de esta información dentro de 2 horas.

---

## 🛠️ Instalación y ejecución

### 1. Instalar dependencias

💡No necesitas tener Go instalado. Pero si deseas correr el código localmente sin contenedores (por ejemplo, para testing o desarrollo fuera de Docker), asegúrate de tener Go instalado (v1.20+).
Luego, en la raíz del proyecto, ejecuta:

```bash
go mod tidy
```

Si deseas correr el proyecto localmente con Go, debes agregar un archivo `.env` en la raíz del proyecto con las variables necesarias para conectar a la base de datos. Te dejamos algunas variables que puedan ser de ayuda:

```
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=nexu
```

### 2. Ejecutar el servidor

Este proyecto está preparado para ejecutarse con Docker. Usa:

```bash
docker-compose up --build
```

Verifica que el mensaje en consola sea:

```
Server running on port 8080
```

Esto indica que el servidor está listo para recibir solicitudes.

### 3. Apagar y limpiar contenedores

Cuando quieras detener el servidor solo presiona el comandos `Ctrl+C`

Para eliminar contenedores y volúmenes asociados:

```bash
docker-compose down -v
```

---

## 🔥 Endpoints disponibles

```http
GET    /brands
GET    /brands/{id}/models
POST   /brands
POST   /brands/{id}/models
PUT    /models/{id}
GET    /models
GET    /models?greater=...&lower=...
```

📬 Colección de Postman
Dentro de la carpeta docs encontrarás una colección de Postman:

```
nexu.postman_collection.json
```

---

## 🧠 Ideas de mejora

- Reestructurar el proyecto siguiendo una **arquitectura hexagonal** (separación de contextos: domain, application, infrastructure).
- Aplicar **patrones de diseño** para desacoplar la base de datos, comandos y mejorar la mantenibilidad.
- Reemplazar `database/sql` por un ORM como [GORM](https://gorm.io/) para una gestión más cómoda de entidades.
- Implementar **migraciones automáticas** desacopladas del arranque (`InitDB`).
- Agregar **middleware de autenticación** (API Key, JWT, etc.) para rutas privadas.
- Crear un servidor/handler más robusto con soporte a middlewares globales personalizados.

---

## 🚧 Tareas pendientes

- Linting: el código debe ser validado con herramientas como `golangci-lint` o `go vet`.
- Agregar pruebas unitarias y de integración (mínimo dos casos de prueba).
- Implementar el cálculo automático de `average_price` por marca. Esto puede lograrse con un **trigger** que reaccione a cambios en la tabla `models`.

---

📅 Última actualización: 1 de abril de 2025