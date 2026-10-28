-- Categorías canónicas de referencia en el mercado ganadero argentino (MAG y Rosgan)
INSERT INTO categorias_canon (codigo, nombre_visible, tipo_destino, peso_minimo_kg, peso_maximo_kg) VALUES
-- FAENA (Mercado Agroganadero de Cañuelas - MAG)
('NOV_PESADO', 'Novillos Pesados', 'FAENA', 430, NULL),
('NOVILLITO', 'Novillitos (Consumo)', 'FAENA', 300, 430),
('VAQUILLONA', 'Vaquillonas (Consumo)', 'FAENA', 270, 400),
('VACA_BUENA', 'Vacas Buenas / Especiales', 'FAENA', 420, NULL),
('VACA_REGULAR', 'Vacas Regulares', 'FAENA', 380, NULL),
('VACA_CONSERVA', 'Vacas Conserva / Manufactura ("China")', 'FAENA', 300, NULL),
('TORO', 'Toros', 'FAENA', 500, NULL),
('MEJ', 'Macho Entero Joven (MEJ)', 'FAENA', 320, 450),

-- INVERNADA (Rosgan y Remates Ferias)
('TERNERO_INV', 'Terneros Invernada (160-180kg)', 'INVERNADA', 140, 200),
('TERNERA_INV', 'Terneras Invernada (150-180kg)', 'INVERNADA', 130, 190),
('TERNERO_A_INV', 'Terneros/as Mixto', 'INVERNADA', 130, 200),
('NOVILLITO_REC', 'Novillitos Recría (200-300kg)', 'INVERNADA', 200, 300),
('VAQUILLONA_REC', 'Vaquillonas Recría (200-280kg)', 'INVERNADA', 200, 280),

-- CRÍA
('VACA_CRIA', 'Vacas con Cría al Pie', 'CRIA', NULL, NULL),
('VACA_PRENADA', 'Vacas Preñadas', 'CRIA', NULL, NULL),
('VAQ_PRENADA', 'Vaquillonas Preñadas', 'CRIA', NULL, NULL)
ON CONFLICT (codigo) DO NOTHING;

-- Mapeos iniciales para Anti-Corruption Layer (MAG y Rosgan)
INSERT INTO mapeos_fuente (fuente, texto_origen, categoria_canon_id, verificado_por_humano)
SELECT 'MAG', 'Novillos', id, true FROM categorias_canon WHERE codigo = 'NOV_PESADO'
UNION ALL
SELECT 'MAG', 'Novillitos', id, true FROM categorias_canon WHERE codigo = 'NOVILLITO'
UNION ALL
SELECT 'MAG', 'Vaquillonas', id, true FROM categorias_canon WHERE codigo = 'VAQUILLONA'
UNION ALL
SELECT 'MAG', 'Vacas Especial', id, true FROM categorias_canon WHERE codigo = 'VACA_BUENA'
UNION ALL
SELECT 'MAG', 'Vacas Regular', id, true FROM categorias_canon WHERE codigo = 'VACA_REGULAR'
UNION ALL
SELECT 'MAG', 'Vacas Conserva Buena', id, true FROM categorias_canon WHERE codigo = 'VACA_CONSERVA'
UNION ALL
SELECT 'MAG', 'Vacas Conserva Inferior', id, true FROM categorias_canon WHERE codigo = 'VACA_CONSERVA'
UNION ALL
SELECT 'MAG', 'Toros', id, true FROM categorias_canon WHERE codigo = 'TORO'
UNION ALL
SELECT 'MAG', 'MEJ', id, true FROM categorias_canon WHERE codigo = 'MEJ'
UNION ALL
SELECT 'ROSGAN', 'Terneros', id, true FROM categorias_canon WHERE codigo = 'TERNERO_INV'
UNION ALL
SELECT 'ROSGAN', 'Terneras', id, true FROM categorias_canon WHERE codigo = 'TERNERA_INV'
UNION ALL
SELECT 'ROSGAN', 'Terneros y terneras', id, true FROM categorias_canon WHERE codigo = 'TERNERO_A_INV'
UNION ALL
SELECT 'ROSGAN', 'Novillitos', id, true FROM categorias_canon WHERE codigo = 'NOVILLITO_REC'
UNION ALL
SELECT 'ROSGAN', 'Vaquillonas', id, true FROM categorias_canon WHERE codigo = 'VAQUILLONA_REC'
UNION ALL
SELECT 'ROSGAN', 'Vacas con cría', id, true FROM categorias_canon WHERE codigo = 'VACA_CRIA'
UNION ALL
SELECT 'ROSGAN', 'Vacas preñadas', id, true FROM categorias_canon WHERE codigo = 'VACA_PRENADA'
ON CONFLICT (fuente, texto_origen) DO NOTHING;

