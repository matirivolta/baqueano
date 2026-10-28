-- name: CheckAlertaEmitida :one
SELECT EXISTS(
    SELECT 1 FROM alertas_log
    WHERE fecha = $1 
      AND categoria_canon_id = $2 
      AND tipo_alerta = $3 
      AND canal = $4
) AS emitida;

-- name: RegistrarAlerta :one
INSERT INTO alertas_log (fecha, categoria_canon_id, tipo_alerta, canal, mensaje, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

