-- name: GetMapeoFuente :one
SELECT id, fuente, texto_origen, categoria_canon_id, verificado_por_humano, created_at
FROM mapeos_fuente
WHERE fuente = $1 AND texto_origen = $2;

-- name: UpsertMapeoFuente :one
INSERT INTO mapeos_fuente (
    fuente, texto_origen, categoria_canon_id, verificado_por_humano
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (fuente, texto_origen)
DO UPDATE SET
    categoria_canon_id = COALESCE(EXCLUDED.categoria_canon_id, mapeos_fuente.categoria_canon_id),
    verificado_por_humano = EXCLUDED.verificado_por_humano
RETURNING *;

