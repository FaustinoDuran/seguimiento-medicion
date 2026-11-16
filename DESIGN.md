---
name: "Seguimiento y medición"
description: "Herramienta operativa oscura inspirada en Linear, sin copiar marca."
colors:
  canvas: "#101113"
  surface: "#17181c"
  surface-muted: "#202228"
  ink: "#eeeff2"
  ink-muted: "#a4a7b3"
  line: "#303239"
  brand: "#5e62d6"
  brand-strong: "#5054be"
  accent: "#b0b3ff"
  accent-soft: "#22233b"
  success: "#86d3af"
  success-soft: "#182b24"
  danger: "#f2a1a9"
  danger-soft: "#321f26"
  control-bg: "#111216"
  control-border: "#626571"
  sidebar: "#0c0d0f"
  table-head: "#18191e"
  table-hover: "#191b21"
  warning: "#e8c9a4"
  warning-soft: "#29221b"
  warning-border: "#65503c"
typography:
  headline:
    fontSize: "24px"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "-0.025em"
    fontFamily: "\"Inter Variable\", Inter, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif"
  title:
    fontSize: "16px"
    fontWeight: 580
    lineHeight: 1.4
    letterSpacing: "-0.015em"
    fontFamily: "\"Inter Variable\", Inter, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif"
  body:
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.5
    fontFamily: "\"Inter Variable\", Inter, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif"
  label:
    fontSize: "12px"
    fontWeight: 500
    fontFamily: "\"Inter Variable\", Inter, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif"
  action:
    fontSize: "12px"
    fontWeight: 550
    fontFamily: '"Inter Variable", Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'
  table:
    fontSize: "13px"
    fontFamily: "\"Inter Variable\", Inter, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif"
  metric:
    fontSize: "24px"
    fontWeight: 500
    lineHeight: 1.3
    fontFamily: "\"Inter Variable\", Inter, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif"
rounded:
  control: "5px"
  panel: "8px"
spacing:
  compact: "8px"
  field-gap: "16px"
  panel: "20px"
  section: "24px"
  page: "32px"
components:
  button-primary:
    backgroundColor: "{colors.brand}"
    textColor: "#fff"
    rounded: "{rounded.control}"
    padding: "7px 12px"
    typography: "{typography.action}"
  button-primary-hover:
    backgroundColor: "{colors.brand-strong}"
  button-link:
    textColor: "{colors.ink}"
    typography: "{typography.label}"
  input:
    backgroundColor: "{colors.control-bg}"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: "8px 10px"
  nav-item:
    textColor: "{colors.ink-muted}"
    rounded: "{rounded.control}"
    padding: "8px 11px"
  nav-item-active:
    backgroundColor: "{colors.surface-muted}"
    textColor: "{colors.ink}"
  status:
    backgroundColor: "{colors.surface-muted}"
    padding: "4px 8px"
  panel:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.panel}"
    padding: "20px"
  panel-mobile:
    padding: "16px"
---
# Design System: Seguimiento y medición

## Overview

**Creative North Star: "Herramienta operativa oscura inspirada en Linear, sin copiar marca"**

Inter Variable local, carbón y capas neutras, acento índigo, bordes y luminancia sin sombras, densidad compacta. La referencia del usuario fija fuentes y colores; la identidad y los SVG geométricos pertenecen al producto.

La lectura prioriza títulos literales, datos y acciones. Texto claro, secundarios apagados y controles nativos mantienen una interfaz directa en español. El tamaño y el espacio distinguen jerarquías dentro de una sola familia sans.

**Key Characteristics:**
- Superficies oscuras con bordes finos.
- Una familia sans local y cifras tabulares en métricas e identificadores.
- Controles compactos, foco visible y adaptación móvil.

## Colors

La paleta combina carbón, neutros de baja saturación e índigo para acciones y orientación. Los valores normativos están en el frontmatter.

### Primary

- **Índigo operativo** (`brand`, `brand-strong`): acción principal y su hover.
- **Lavanda clara** (`accent`, `accent-soft`): foco, selección de pestaña y apoyos del acento.

### Neutral

- **Carbón base** (`canvas`, `sidebar`): espacio de trabajo y navegación persistente.
- **Capas de trabajo** (`surface`, `surface-muted`, `control-bg`): paneles, estado activo, etiquetas y campos.
- **Texto claro y secundario** (`ink`, `ink-muted`): contenido y contexto subordinado.
- **Líneas de estructura** (`line`, `control-border`): separación de regiones y reconocimiento de campos.
- **Filas discretas** (`table-head`, `table-hover`): encabezado y seguimiento de filas.

Los colores semánticos `success`/`success-soft` y `danger`/`danger-soft` distinguen resultados y errores. `warning`, `warning-soft` y `warning-border` corresponden a avisos; no son un segundo acento de navegación. Las barras distinguen planificación en índigo y finalización en verde.

## Typography

**Display Font:** Inter Variable con la misma cadena de respaldo sans que el cuerpo.
**Body Font:** Inter Variable alojada en `/fonts/InterVariable.woff2`, con licencia OFL en el mismo directorio.

**Character:** sans legible y contenida. El peso variable diferencia niveles sin cambiar de familia; no hay escala de portada.

### Hierarchy

- **Headline:** título principal; tracking ligeramente cerrado.
- **Title:** encabezado de sección.
- **Body:** base de la aplicación; tablas y descripción operativa usan la escala `table`.
- **Label:** etiquetas y acciones; encabezados de tabla descienden a (11px), peso (500), sin mayúsculas forzadas.
- **Metric:** valor numérico con `tabular-nums`; las unidades y el contexto quedan subordinados.

**The Single Family Rule.** Usar Inter Variable para títulos, texto y cifras destacadas; el título principal conserva la escala compacta.

## Layout

Shell en dos columnas: rail fijo visual de (212px) y contenido flexible. El rail permanece pegado al borde superior y ocupa la altura visible. El contenido tiene ancho máximo de (1280px) y margen horizontal automático. El espacio lateral usa `page`; pasa a (24px) bajo (880px) y (16px) bajo (640px).

Bajo (880px), el rail mide (172px). Bajo (640px), se convierte en barra superior de (56px), se oculta la copia de marca y los iconos de navegación; los enlaces conservan texto. Formularios y métricas ajustan columnas según su contenido. La grilla de proyecto utiliza cuatro columnas desde (1100px), dos en el tramo intermedio y una en móvil.

El ritmo reutiliza separaciones compactas, de campo, panel y sección. Las tablas mantienen región de desplazamiento horizontal; la lista de proyectos adopta filas con etiquetas en móvil. No se fuerza a las tablas de trabajo extensas a caber comprimiendo su contenido. Identificadores permanecen unidos y los títulos pueden envolver.

## Elevation & Depth

Los paneles no usan sombras. El fondo, los contenedores, los estados activos y las líneas producen profundidad por luminancia. El hover cambia fondo o color; el foco es un contorno visible, no elevación.

**The Luminance Rule.** Separar regiones con capas neutras y bordes; los paneles no usan sombras.

## Shapes

Controles y navegación comparten esquina corta `control`; los paneles y regiones de tabla usan `panel`. Bordes sólidos de (1px) delimitan estructura. Los estados vacíos usan borde discontinuo. Las etiquetas de estado son rectangulares, con relleno compacto. Marca e iconos son SVG geométricos; no se incorporó imagen raster de producto.

## Components

### Buttons

Acciones compactas. La variante principal usa índigo, texto blanco, altura mínima de (34px), borde transparente y peso (550). Su hover oscurece el fondo durante (140ms ease). El estado deshabilitado baja opacidad a (.55) e indica espera. Los enlaces de acción usan texto claro; su hover incorpora lavanda y subrayado. Las acciones discretas de módulos conservan subrayado y fondo transparente.

### Chips

Etiquetas informativas de estado, con superficie neutra y relleno `status`; los estados completados o abiertos de sprint usan fondo verde oscuro. No actúan como botones ni tienen hover interactivo.

### Cards / Containers

Paneles de formulario y medición con borde fino, superficie elevada por luminancia y radio `panel`. El relleno normal usa `panel`; en móvil baja a (16px). Las métricas se agrupan en una región continua con divisiones y cifras tabulares.

### Inputs / Fields

Campos nativos de fondo oscuro, texto claro y borde de mayor contraste que la línea estructural. Altura mínima de (38px), tipografía de (13px). Las áreas de texto permiten redimensionado vertical. Errores cambian el borde al color semántico y conservan mensaje visible.

El foco general usa contorno lavanda de (2px) separado (3px). Hay variaciones existentes: medición usa contorno claro separado (2px) y el selector de reporte usa lavanda de (3px). Son diferencias locales, no una nueva regla para futuras superficies.

### Navigation

Enlaces del rail con altura mínima de (38px), tipografía de (13px) y secundario apagado. Hover aclara texto y añade superficie; activo usa la capa neutra superior y peso (550). Las pestañas de proyecto usan texto de (12px), altura mínima de (40px), línea inferior lavanda en activo y salto de línea cuando falta espacio. En móvil alcanzan (42px).

### Tables

Texto operativo, encabezados pequeños y líneas finas. Celdas con relleno de (12px 14px); el hover de fila aporta luminancia discreta. Los títulos conservan flexibilidad y los identificadores usan cifras tabulares. La región desplazable mantiene el ancho dentro del contenido.

## Do's and Don'ts

### Do:
- **Do** conservar los tokens del frontmatter como referencia normativa.
- **Do** mantener controles nativos, etiquetas visibles y foco de teclado.
- **Do** permitir que títulos largos se ajusten y que tablas extensas se desplacen dentro de su región.

### Don't:
- **Don't** copiar la marca ni el contenido comercial de Linear.
- **Don't** agregar sombras o brillos a los paneles planos.
- **Don't** introducir una familia serif o titulares promocionales en el espacio operativo.
