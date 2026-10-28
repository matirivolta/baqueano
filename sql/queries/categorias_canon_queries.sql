-- name: ListCategoriasCanon :many
SELECT id, codigo, nombre_visible, tipo_destino, peso_minimo_kg, peso_maximo_kg, activo, created_at, updated_at
FROM categorias_canon
WHERE activo = true
ORDER BY tipo_destino, nombre_visible;

-- name: GetCategoriaCanonByCodigo :one
SELECT id, codigo, nombre_visible, tipo_destino, peso_minimo_kg, peso_maximo_kg, activo, created_at, updated_at
FROM categorias_canon
WHERE codigo = $1;

-- name: GetCategoriaCanonByID :one
SELECT id, codigo, nombre_visible, tipo_destino, peso_minimo_kg, peso_maximo_kg, activo, created_at, updated_at
FROM categorias_canon
WHERE id = $1;

