CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Categorías canónicas
CREATE TABLE categorias_canon (
    id SERIAL PRIMARY KEY,
    codigo VARCHAR(30) UNIQUE NOT NULL,
    nombre_visible VARCHAR(80) NOT NULL,
    tipo_destino VARCHAR(20) NOT NULL CHECK (tipo_destino IN ('INVERNADA', 'FAENA', 'CRIA')),
    peso_minimo_kg INT,
    peso_maximo_kg INT,
    activo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Mapeos para el Anti-Corruption Layer
-- Nota: categoria_canon_id es NULLable para permitir el registro de términos desconocidos
-- pendientes de homologación humana sin violar la integridad referencial.
CREATE TABLE mapeos_fuente (
    id SERIAL PRIMARY KEY,
    fuente VARCHAR(50) NOT NULL,
    texto_origen VARCHAR(150) NOT NULL,
    categoria_canon_id INT REFERENCES categorias_canon(id) ON DELETE RESTRICT,
    verificado_por_humano BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unq_fuente_texto UNIQUE(fuente, texto_origen)
);

CREATE INDEX idx_mapeos_fuente_lookup ON mapeos_fuente(fuente, texto_origen);

-- Cotizaciones normalizadas
CREATE TABLE cotizaciones (
    id BIGSERIAL PRIMARY KEY,
    fecha DATE NOT NULL,
    fuente VARCHAR(50) NOT NULL,
    categoria_canon_id INT NOT NULL REFERENCES categorias_canon(id) ON DELETE RESTRICT,
    precio_min NUMERIC(12,2),
    precio_max NUMERIC(12,2),
    precio_promedio NUMERIC(12,2) NOT NULL,
    volumen_cabezas INT DEFAULT 0,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unq_fecha_fuente_cat UNIQUE(fecha, fuente, categoria_canon_id)
);

CREATE INDEX idx_cotizaciones_fecha_cat ON cotizaciones(fecha DESC, categoria_canon_id);
CREATE INDEX idx_cotizaciones_cat_fecha ON cotizaciones(categoria_canon_id, fecha DESC);

-- Registro de auditoría de alertas despachadas
CREATE TABLE alertas_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fecha DATE NOT NULL,
    categoria_canon_id INT NOT NULL REFERENCES categorias_canon(id),
    tipo_alerta VARCHAR(40) NOT NULL,
    canal VARCHAR(20) NOT NULL DEFAULT 'TELEGRAM',
    mensaje TEXT NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    disparado_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alertas_idempotencia ON alertas_log(fecha, categoria_canon_id, tipo_alerta, canal);

