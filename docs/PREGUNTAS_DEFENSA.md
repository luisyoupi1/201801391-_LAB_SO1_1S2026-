# Preguntas para la defensa

## ¿Por qué son tres runtimes diferentes?

El objetivo es demostrar que una imagen OCI puede distribuirse mediante un registro estándar y ejecutarse con herramientas distintas. Containerd ejecuta API1/API2 en VM1, Podman ejecuta API3 en VM2 y Docker administra Zot en VM3.

## ¿Qué hace nerdctl?

Containerd no ofrece por sí solo una experiencia de construcción y ejecución equivalente a Docker. nerdctl es un cliente compatible con flujos de construcción, `push`, `pull` y ejecución sobre Containerd.

## ¿Cómo verifica una API a otra?

El endpoint `call-api#` hace un `GET` real al `/health` de la API destino. Solo devuelve `connection:true` cuando hay HTTP 2xx, JSON válido y `status:"UP"`.

## ¿Qué pasa si una API se cae?

El cliente HTTP tiene timeout. Una conexión rechazada, timeout, respuesta no 2xx, JSON inválido o estado distinto de `UP` produce `connection:false` y el mensaje `ERROR` requerido, sin derribar la API que hizo la consulta.

## ¿Por qué se usa red host?

Las APIs se distribuyen entre runtimes distintos y dos VMs. La red host mantiene puertos e IPs del laboratorio explícitos y evita depender de una red virtual específica de Docker, Podman o Containerd.

## ¿Por qué el registro se configura como inseguro?

Solo para el segmento privado y aislado `192.168.56.0/24`, donde Zot usa HTTP para simplificar el laboratorio. No es una práctica de producción; fuera del laboratorio debe habilitarse TLS y autenticación.

## ¿Por qué los nombres de imagen están en minúscula?

El formato pedido es `API#-#CARNET`, pero los repositorios OCI/Docker requieren nombres en minúscula. Por eso se conserva la estructura semántica como `api1-201801391:1.0.0`.

## ¿Dónde se cambia el carnet?

La variable `CARNET` controla rutas y respuestas. El valor predeterminado ya es `201801391`. Los scripts de despliegue y las imágenes usan la misma variable para evitar inconsistencias.

## ¿Qué cambiarías en producción?

TLS y autenticación en Zot, secretos gestionados fuera del repositorio, redes segmentadas, observabilidad, límites de recursos, versiones de imágenes fijadas por digest, réplicas y un orquestador como Kubernetes.

## Cambio sencillo que puedes demostrar en vivo

Modifica el timeout del cliente en `internal/service/server.go`, ejecuta `go test ./...` y explica que un timeout menor detecta fallos más rápido pero tolera menos latencia. Otra opción es cambiar el puerto predeterminado en el `main.go` de una API y actualizar su variable `PORT` en el despliegue.
