# Auditoría de seguridad del enunciado

Fuente revisada: `Proyecto1-SO1-2doSem.docx.pdf`
SHA-256: `F908E797D760A06687D6DFB50F9CA43FD9A6850DE08759A1940E9CA62446051C`

## Resultado

No se detectaron instrucciones ocultas ni contenido activo malicioso. Los requisitos implementados en este repositorio corresponden al texto visible del documento.

## Revisiones realizadas

- Renderizado e inspección visual de las 16 páginas.
- Extracción y comparación del texto de todas las páginas.
- Búsqueda de texto blanco, tamaño menor a 4 puntos, texto fuera de la página y modo de renderizado invisible.
- Revisión de JavaScript, acciones automáticas, archivos incrustados, formularios, acciones de lanzamiento y contenido multimedia.
- Inspección de anotaciones y enlaces.

No se encontraron formularios, JavaScript, acciones de apertura, adjuntos, texto invisible, texto blanco, texto diminuto ni texto fuera del área visible. Hay dos anotaciones URI visibles que apuntan al enlace de apoyo sobre runtimes mostrado en la página 10. El lector reportó una clave interna duplicada en un diccionario PDF; el archivo fue producido por el renderizador de Google Docs y el hallazgo no representa una instrucción ni una acción ejecutable.
