# Guía para defensa

## ¿Por qué se usa `seq_file`?

Porque la lista de procesos puede superar una página. `seq_file` maneja lecturas parciales,
offsets y buffers sin construir manualmente una interfaz `/proc` frágil.

## ¿Cómo se evita usar un `task_struct` liberado?

RCU solo se mantiene para copiar PIDs. Después se obtiene una referencia con
`get_pid_task`; al terminar se llama `put_task_struct`.

## Diferencia entre VSZ y RSS

VSZ es el espacio virtual reservado por el proceso. RSS son las páginas físicas residentes
actualmente en RAM. Un proceso puede tener VSZ alto y RSS bajo.

## ¿Cómo se relaciona un PID con Docker?

Se lee `/proc/<pid>/cgroup` y se busca el ID completo o corto del contenedor retornado por
`docker inspect`. Si no aparece, se compara el PID con `State.Pid`.

## ¿Por qué eBPF y no confiar en la respuesta de Docker?

La respuesta de Docker confirma una operación de alto nivel. El tracepoint confirma que el
kernel recibió realmente una llamada `kill(2)` con PID y señal específicos.

## ¿Cómo se garantiza que Grafana no sea eliminado?

El selector solo considera contenedores con `so1.project=201801391`. La infraestructura no
lleva esa etiqueta. Además, las decisiones solo aceptan perfiles `low`, `high` o `intruder`.

## ¿Cómo se mantienen los mínimos?

Primero se agrupa por perfil. Solo se selecciona el excedente sobre 3 bajos o 2 altos. Al
final del ciclo, el generador restaura cualquier faltante.

## ¿Qué pasa si no llega confirmación eBPF?

El contenedor puede detenerse y retirarse, pero no incrementa el contador de eliminaciones
confirmadas en Valkey. El timeout queda en el journal para diagnóstico.

## ¿Por qué Valkey y Prometheus?

Valkey conserva instantáneas, históricos y eventos como fuente de registro. Prometheus
recoge una proyección numérica del daemon que Grafana puede consultar eficientemente.

## ¿Cómo se apaga limpiamente?

Systemd envía `SIGTERM`. El contexto cancela lectores y servidor, cierra Valkey y eBPF, y
retira el cron instalado.

