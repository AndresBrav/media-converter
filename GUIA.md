# Guía de Uso: Media Converter CLI (Go)

`media-converter` es una herramienta de línea de comandos (CLI) escrita en **Go puro** (sin dependencias de CGo/GCC) diseñada para convertir imágenes en lote de forma concurrente mediante un pool de trabajadores (workers).

---

## 1. Métodos de Ejecución

Existen dos formas principales de correr y probar la herramienta:

### Método A: Ejecución directa en desarrollo (Recomendado para pruebas rápidas)

Permite compilar y ejecutar el código fuente en un solo paso, sin dejar un archivo `.exe` en tu carpeta:
bash
go run . --input .\imagenes --output .\resultado --format webp --quality 80

_También puedes abreviarlo usando flags cortos:_
bash
go run . -i .\imagenes -o .\resultado -f webp -q 80

### Método B: Compilación y distribución (Generando el binario ejecutable)

1. **Compilar el proyecto:** Genera el archivo ejecutable `media-converter.exe` en tu carpeta:
   bash
   go build -o media-converter.exe main.go
2. **Ejecutar el binario generado:**
   bash
   .\media-converter.exe --input .\imagenes --output .\resultado --format webp --quality 80

---

## 2. Parámetros del Comando (Flags)

| Parámetro (Largo / Corto) | Tipo     | Obligatorio | Descripción / Valores                                                                                                                 |
| :------------------------ | :------- | :---------- | :------------------------------------------------------------------------------------------------------------------------------------ |
| `--input` / `-i`          | `string` | **Sí**      | Ruta de la carpeta que contiene las imágenes originales a procesar.                                                                   |
| `--output` / `-o`         | `string` | **Sí**      | Ruta de la carpeta donde se guardarán las imágenes convertidas (se crea automáticamente si no existe).                                |
| `--format` / `-f`         | `string` | **Sí**      | Formato de destino. Soporta: `png`, `jpg`, `jpeg`, `webp` (insensible a mayúsculas).                                                  |
| `--workers` / `-w`        | `int`    | No          | Número de hilos/workers paralelos. **Por defecto:** Número de núcleos de tu CPU.                                                      |
| `--quality` / `-q`        | `int`    | No          | Calidad de compresión (rango de `1` a `100`). **Por defecto:** `80`. _(Aplica para WebP y JPEG; se ignora en PNG ya que es lossless)_ |

---

## 3. Ejemplos Prácticos

- **Ejemplo 1: Conversión estándar a WebP (Calidad optimizada por defecto al 80%)**
  bash
  go run . --input .\imagenes --output .\resultado --format webp
- **Ejemplo 2: Conversión a JPEG con compresión fuerte (60% de calidad para ahorrar espacio)**
  bash
  go run . -i .\imagenes -o .\resultado -f jpeg -q 60
- **Ejemplo 3: Forzando a usar exactamente 4 workers paralelos**
  bash
  go run . -i .\imagenes -o .\resultado -f webp -w 4 -q 85

---

## 4. Consola de Monitoreo y Logs

Al ejecutar la herramienta verás un reporte detallado en tiempo real estructurado de la siguiente forma:

1. **Configuración inicial**: Muestra las carpetas de entrada/salida, formato y workers lanzados.
2. **Logs de Inicio de Workers**: Cada worker avisa cuando se activa de forma concurrente (`Worker X started`).
3. **Contador de Progreso en Tiempo Real**: Cada conversión muestra su número de tarea procesada sobre el total:
    - Exitosa: `[1/16] ✓ imagen.jpg -> imagen.webp`
    - Fallida: `[2/16] ✗ Error convirtiendo foto.png: [detalle de error]`
4. **Resumen Final**:
    - `Errores: X` (Conteo total de archivos corruptos o incompatibles).
    - `Total: X.Xs` (Tiempo total exacto de procesamiento).

### Captura de ejemplo en consola:

text
Configuration

---

Input : .\imagenes
Output : .\resultado
Format : .webp
Workers: 2
Resume
imagenes\Llama.webp -> resultado\Llama.webp
imagenes\astronaut.jpg -> resultado\astronaut.webp
Archivos encontrados: 2
Lanzando 2 workers...

Worker 2 started
Worker 1 started
[1/2] ✓ Llama.webp -> Llama.webp
[2/2] ✓ astronaut.jpg -> astronaut.webp

Todos los trabajos completados
Errores: 0
Total: 0.8s

---

### Referencias del código fuente:

- [cmd/root.go](cmd/root.go)
- [converter/worker.go](converter/worker.go)
- [converter/converter.go](converter/converter.go)
