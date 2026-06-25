# 🎬 Media Converter CLI — Guía de Uso

`media-converter` es una herramienta de línea de comandos (CLI) escrita en **Go** para convertir imágenes y videos en lote de forma concurrente mediante un pool de trabajadores (workers).

---

## 1. Métodos de Ejecución

### Método A: Ejecución directa en desarrollo (Recomendado para pruebas rápidas)

Compila y ejecuta el código en un solo paso sin generar un `.exe`:

```bash
go run . --input .\imagenes --output .\resultado --format webp --quality 80
```

También puedes usar flags cortos:

```bash
go run . -i .\imagenes -o .\resultado -f webp -q 80
```

### Método B: Compilar y distribuir (Generando el binario ejecutable)

1. **Compilar el proyecto:**
   ```bash
   go build -o media-converter.exe .
   ```
2. **Ejecutar el binario:**
   ```bash
   .\media-converter.exe --input .\imagenes --output .\resultado --format webp --quality 80
   ```

---

## 2. Flags Disponibles

| Flag | Corto | Tipo | Default | Obligatorio | Descripción |
|------|-------|------|---------|:-----------:|-------------|
| `--input` | `-i` | string | — | ✅ Sí | Carpeta con los archivos originales a procesar |
| `--output` | `-o` | string | — | ✅ Sí | Carpeta de destino (se crea automáticamente si no existe) |
| `--format` | `-f` | string | — | ✅ Sí | Formato de salida deseado |
| `--workers` | `-w` | int | núcleos CPU | No | Número de workers paralelos para el procesamiento concurrente |
| `--quality` | `-q` | int (1-100) | `80` | No | Calidad de compresión (aplica en JPEG y WebP; se ignora en PNG) |
| `--width` | — | int | `0` | No | Ancho máximo en píxeles (`0` = mantener original) |
| `--height` | — | int | `0` | No | Alto máximo en píxeles (`0` = mantener original) |
| `--watermark` | — | string | — | No | Ruta a una imagen que se usará como marca de agua |
| `--thumbnail` | — | bool | `false` | No | Genera una miniatura adicional de 150×150 px por cada archivo |
| `--recursive` | `-r` | bool | `false` | No | Procesa directorios y subcarpetas de manera recursiva |

---

## 3. Formatos Soportados

### 🖼️ Imágenes

| Entrada | Salida |
|---------|--------|
| `.jpg` / `.jpeg` | `.jpg` / `.jpeg` |
| `.png` | `.png` |
| `.webp` | `.webp` |

La conversión puede hacerse entre cualquier combinación: `jpg → webp`, `png → jpg`, `webp → png`, etc.

### 🎥 Videos

| Entrada | Salida |
|---------|--------|
| `.mkv` | `.mp4` |
| `.avi` | `.mp4` |
| `.mov` | `.mp4` |
| `.mp4` | `.mp4` |

> ⚠️ La conversión de video **requiere FFmpeg instalado** y disponible en el PATH.
> Puedes instalarlo con:
> ```bash
> winget install ffmpeg
> ```
> O descargarlo desde: https://ffmpeg.org/download.html

---

## 4. Funcionamiento del Procesamiento Recursivo (`-r`)

Cuando se activa el flag `--recursive` o `-r`, el programa cambia su comportamiento de búsqueda:

1. **Búsqueda exhaustiva:** Utiliza `filepath.Walk` para recorrer el directorio de entrada y todas sus subcarpetas en profundidad.
2. **Preservación de la estructura:** Mide la ruta relativa de cada archivo respecto al directorio de entrada original. Recrea exactamente esa misma estructura de carpetas dentro del directorio de salida.
3. **Creación dinámica:** Si una subcarpeta no existe en el destino, el programa la crea automáticamente (`os.MkdirAll`) antes de delegar la conversión a los workers.

### Ejemplo de estructura

**Entrada:**
```text
📂 imagenes/
 ├── 📷 foto1.jpg
 └── 📂 vacaciones/
      └── 📷 foto2.png
```

**Comando:**
```bash
go run . -i .\imagenes -o .\resultado -f webp -r
```

**Salida generada:**
```text
📂 resultado/
 ├── 📷 foto1.webp
 └── 📂 vacaciones/
      └── 📷 foto2.webp
```

---

## 5. Funcionamiento del Pool de Workers Concurrentes (`-w`)

El procesamiento no se realiza de forma secuencial (archivo por archivo), sino utilizando un modelo de concurrencia basado en **Worker Pools**:

1. **Planificación (Jobs):** El programa escanea el directorio de entrada (plano o recursivo) y genera una lista de tareas (`Job`).
2. **Canal de comunicación (Channel):** Se crea un canal en Go donde se depositan todos los trabajos a realizar.
3. **Trabajadores (Workers):** Se lanzan de forma simultánea `N` hilos de ejecución (por defecto, el número de núcleos lógicos del procesador, o el valor indicado con `-w`).
4. **Consumo concurrente:** Todos los workers escuchan el mismo canal. En cuanto un worker termina de procesar una imagen o video, toma inmediatamente el siguiente trabajo disponible en el canal.
5. **Sincronización:** Un grupo de espera (`sync.WaitGroup`) coordina a los workers para asegurar que el programa principal no finalice hasta que la totalidad de los trabajos del canal hayan sido completados.

---

## 6. Ejemplos Prácticos

### Conversión estándar y concurrencia

**Conversión básica con workers por defecto (núcleos de CPU)**
```bash
go run . -i .\imagenes -o .\resultado -f webp
```

**Limitar el procesamiento a exactamente 2 workers paralelos**
```bash
go run . -i .\imagenes -o .\resultado -f webp -w 2
```

**Convertir a JPEG con menor calidad (ahorra espacio)**
```bash
go run . -i .\imagenes -o .\resultado -f jpeg -q 60
```

### Procesamiento recursivo

**Procesar recursivamente todas las subcarpetas manteniendo la estructura en el destino**
```bash
go run . -i .\imagenes -o .\resultado -f webp -r
```

**Procesar recursivamente limitando el uso de recursos a 4 workers**
```bash
go run . -i .\imagenes -o .\resultado -f webp -r -w 4
```

### Edición y Filtros

**Redimensionar a un ancho fijo (mantiene proporción)**
```bash
go run . -i .\imagenes -o .\resultado -f jpg --width 800
```

**Redimensionar con ancho Y alto fijos (recorta y centra)**
```bash
go run . -i .\imagenes -o .\resultado -f webp --width 1280 --height 720
```

**Agregar marca de agua en la esquina inferior derecha**
```bash
go run . -i .\imagenes -o .\resultado -f jpg --watermark .\logo.png
```

**Generar miniaturas de 150×150 junto a cada imagen convertida**
```bash
go run . -i .\imagenes -o .\resultado -f webp --thumbnail
```

**Comando completo combinado (recursivo, redimensionado, marca de agua y miniaturas)**
```bash
go run . -i .\imagenes -o .\resultado -f webp -r -w 4 --width 1280 --height 720 --watermark .\logo.png --thumbnail
```

### Conversión de videos

**Convertir videos a MP4 recursivamente**
```bash
go run . -i .\videos -o .\resultado -f mp4 -r
```

---

## 7. Consola de Monitoreo

Al ejecutar la herramienta verás un reporte en tiempo real:

```text
Configuration
---
Input   : .\imagenes
Output  : .\resultado
Format  : .webp
Workers : 4

Resume
imagenes\foto1.jpg -> resultado\foto1.webp
imagenes\vacaciones\foto2.png -> resultado\vacaciones\foto2.webp
Archivos encontrados: 2
Lanzando 4 workers...

Worker 1 started
Worker 2 started
Worker 3 started
Worker 4 started
[1/2] ✓ foto1.jpg -> foto1.webp
[2/2] ✓ vacaciones\foto2.png -> vacaciones\foto2.webp

Todos los trabajos completados
Errores: 0
Total: 0.8s
```

- ✓ = Conversión exitosa
- ✗ = Error (se muestra el detalle del error)

---

## 8. Referencias del Código Fuente

- [cmd/root.go](cmd/root.go) — Definición de flags y lógica principal del CLI
- [converter/converter.go](converter/converter.go) — Conversión de imágenes y videos
- [converter/validator.go](converter/validator.go) — Validación de formatos y directorios
- [converter/job.go](converter/job.go) — Estructura de trabajos, lectura de directorios y caminata recursiva (`Walk`)
- [converter/worker.go](converter/worker.go) — Pool de workers concurrentes
- [converter/logger.go](converter/logger.go) — Logs de progreso en consola
