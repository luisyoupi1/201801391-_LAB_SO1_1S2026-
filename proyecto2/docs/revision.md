# Revisión funcional y de presentación

## Cambios

- El Makefile usa Go 1.24 de Ubuntu cuando está disponible y permite elegir otro compilador con `GO`.
- El instalador incluye GCC 12 y Go 1.24 y conserva Docker si ya está instalado.
- El módulo declara Go 1.20 como versión mínima, requerida por la dependencia eBPF.
- Las duraciones nulas o negativas se rechazan antes de iniciar el servicio.
- Las etiquetas Prometheus se escapan una sola vez. Los contenedores detenidos ya no se presentan como series actuales.
- La marca de última lectura empieza en cero cuando aún no hay muestras.
- El archivo C de eBPF se excluye de la compilación Go mediante una etiqueta de compilación.
- El dashboard organiza 17 paneles, incluidas sus secciones, con disponibilidad, antigüedad de lecturas, memoria y barras históricas de CPU/RAM. Las barras muestran porcentajes reales en lugar de repartir el top 5 como un pastel.

## Verificación realizada

- Compilación del módulo del kernel, objeto eBPF y daemon Go.
- `go test -race ./...` y `go vet ./...` satisfactorios.
- Pruebas de regresión para duraciones, etiquetas de métricas y estado inicial.
- Validación de estructura, métricas requeridas e inexistencia de paneles superpuestos.
- Sintaxis de todos los scripts Bash correcta.
- Configuración de Prometheus validada con `promtool` de la imagen del proyecto.
- Grafana 12.1.1 inició y provisionó el dashboard; su API devolvió los 17 paneles.

## Pendiente de prueba en el host

El servicio estaba inactivo y sudo requería contraseña. No se verificaron la carga del módulo, la conexión eBPF al kernel, la generación por cron, el ciclo real de eliminación ni la persistencia completa en Valkey. La compilación y las pruebas anteriores no sustituyen esa verificación de extremo a extremo.

Desde `proyecto2`, después de revisar la configuración:

```bash
cp -n .env.example .env
make all
make test
sudo make install
```

La instalación inicia el servicio y sus cargas de trabajo. Las comprobaciones posteriores están en `docs/guia-instalacion.md`.

## Corrección de confirmaciones eBPF

La ejecución inicial confirmó métricas, Prometheus en UP, Grafana, cron y persistencia en Valkey. Se observaron tiempos de espera al confirmar eliminaciones. La sonda original observaba únicamente `sys_enter_kill`, incluidos sondeos de señal cero; esa entrada no demuestra la generación de una señal de terminación.

Se conserva la auditoría de `sys_enter_kill` y se agrega `signal_generate` con PID del host, filtrando SIGTERM/SIGKILL y resultado aceptado. El envío puede originarse en otras llamadas o espacios de PID. La confirmación se entrega antes de persistir el evento, y las esperas se cancelan al fallar o vencer. La eliminación solo se cuenta después de que Docker la retire y Valkey registre el resultado.

Referencia de la estructura y resultados del tracepoint: [Linux 6.8, signal.h](https://github.com/torvalds/linux/blob/v6.8/include/trace/events/signal.h). Una señal generada no demuestra por sí sola que el proceso haya terminado.

La corrección compiló y pasó `go test -race ./...`, `go vet ./...` y validación estructural. Pendiente: instalarla y verificar que el kernel acepte la nueva sonda y que confirme las eliminaciones reales. `sudo make install` ahora reinicia el servicio para aplicar las actualizaciones.

## Eliminación automática de Docker

Después de reinstalar, la nueva sonda cargó y registró eventos `signal_generate`. Se detectó una carrera entre `docker stop` y `docker rm` porque las cargas usan `--rm`. El daemon ahora verifica que el contenedor haya desaparecido mediante una consulta exitosa a Docker y espera hasta cinco segundos si la eliminación ya está en curso. Los errores de conexión y permisos se mantienen como errores. Compilación, pruebas de carrera y análisis estático pasaron. Falta reinstalar esta corrección y observar el contador de eliminaciones.

## Verificación posterior a la instalación corregida

Se verificó el servicio activo y un ciclo con cuatro eliminaciones confirmadas. Valkey contiene evidencia de una eliminación con origen `signal_generate` y señal 9. Al terminar el ciclo quedaron tres contenedores low y dos high. Prometheus reportó el target UP sin errores y Grafana respondió con base de datos OK. Esta comprobación resuelve los pendientes anteriores de carga de la sonda y registro de eliminaciones; no constituye una prueba prolongada de estabilidad.
