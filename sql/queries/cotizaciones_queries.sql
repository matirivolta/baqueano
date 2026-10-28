-- name: UpsertCotizacion :one
INSERT INTO cotizaciones (
    fecha, fuente, categoria_canon_id, precio_min, precio_max, precio_promedio, volumen_cabezas, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (fecha, fuente, categoria_canon_id) 
DO UPDATE SET
    precio_min = EXCLUDED.precio_min,
    precio_max = EXCLUDED.precio_max,
    precio_promedio = EXCLUDED.precio_promedio,
    volumen_cabezas = EXCLUDED.volumen_cabezas,
    metadata = EXCLUDED.metadata,
    created_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetUltimasCotizacionesPorCategoria :many
SELECT DISTINCT ON (cot.categoria_canon_id, cot.fuente)
    cot.id,
    cot.fecha,
    cot.fuente,
    cot.categoria_canon_id,
    c.codigo AS categoria_codigo,
    c.nombre_visible AS categoria_nombre,
    c.tipo_destino,
    cot.precio_promedio,
    cot.precio_min,
    cot.precio_max,
    cot.volumen_cabezas
FROM cotizaciones cot
JOIN categorias_canon c ON cot.categoria_canon_id = c.id
WHERE c.activo = true
ORDER BY cot.categoria_canon_id, cot.fuente, cot.fecha DESC;

-- name: GetSerieHistorica :many
SELECT fecha, precio_promedio, precio_min, precio_max, volumen_cabezas, fuente
FROM cotizaciones
WHERE categoria_canon_id = $1 AND fecha >= $2
ORDER BY fecha ASC;

-- name: GetCotizacionPrevia :one
SELECT id, fecha, fuente, categoria_canon_id, precio_min, precio_max, precio_promedio, volumen_cabezas, metadata, created_at
FROM cotizaciones
WHERE categoria_canon_id = $1 AND fuente = $2 AND fecha < $3
ORDER BY fecha DESC
LIMIT 1;

