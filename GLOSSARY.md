# Glosario de Dominio - Baqueano (Lenguaje Ubicuo)

Este glosario formaliza los términos y conceptos del dominio ganadero argentino utilizados en la arquitectura y código de **Baqueano**.

---

## 1. Categorías y Destinos de Hacienda

*   **Categoría Canónica (`categorias_canon`):** Clasificación normalizada y unificada por el sistema (ej. `NOVILLITO`, `TERNERO_INV`, `VACA_CONSERVA`) independiente del texto utilizado por cada mercado o consignataria.
*   **Destino (`tipo_destino`):**
    *   **FAENA:** Hacienda lista para sacrificio y procesamiento frigorífico (consumo interno o exportación). Principal plaza formadora: Mercado Agroganadero de Cañuelas (MAG).
    *   **INVERNADA:** Animales jóvenes destinados a la recría y engorde a campo o en feedlot (terneros, terneras, novillitos de recría). Principal plaza: Rosgan y remates ferias del interior.
    *   **CRÍA:** Vientres reproductores (vacas y vaquillonas preñadas o vacas con cría al pie) para la producción de terneros en establecimientos de cría.
*   **Anti-Corruption Layer (ACL) / Mapeos Fuente (`mapeos_fuente`):** Mecanismo de traducción que mapea las cadenas heterogéneas devueltas por los scrapers de las fuentes externas al identificador de la categoría canónica correspondiente. Si un término no existe, queda registrado con categoría nula para homologación supervisada.

---

## 2. Comercialización y Formación de Precios

*   **Cotización (`cotizaciones`):** Registro diario de precios observado en una plaza para una categoría. Incluye:
    *   `precio_promedio`: Precio ponderado representativo del día ($/kg vivo).
    *   `precio_min` y `precio_max`: Rango de precios extremos alcanzados en subasta o remate.
    *   `volumen_cabezas`: Cantidad total de cabezas operadas.
*   **MAG (Mercado Agroganadero de Cañuelas):** Mercado concentrador físico de referencia nacional para hacienda con destino a faena.
*   **Rosgan:** Mercado ganadero televisado de la Bolsa de Comercio de Rosario, referencia nacional de precios de invernada y cría.
*   **Consignataria:** Casa intermediaria que comercializa la hacienda por cuenta y orden del productor, percibiendo una comisión sobre el valor bruto negociado.

---

## 3. Liquidación y Cobro Neto

*   **Desbaste (%):** Merma porcentual pactada sobre el peso bruto del animal para compensar el llenado de panza/agua previo al pesaje en campo (típicamente entre 2% y 8%). En balanza pública o destino suele ser 0%.
*   **Kilos Netos:** Kilos efectivos facturados tras aplicar el desbaste:
    $$Kg_{\text{netos}} = Kg_{\text{brutos}} \times \left(1 - \frac{\text{desbaste}\%}{100}\right)$$
*   **Importe Bruto Hacienda:** Monto liquidado por los animales:
    $$\text{Importe Bruto} = Kg_{\text{netos}} \times \text{Precio}_{\$/kg}$$
*   **Comisión Consignataria:** Porcentaje pactado (usualmente 2% al 3%) sobre el Importe Bruto. Devenga IVA sobre servicios (alícuota 21%).
*   **Flete Jaula:** Costo de traslado por camión jaula desde el establecimiento rural hasta el destino final (cotizado por km o suma fija por jaula/viaje).
*   **Gastos Sanitarios y Administrativos:** Tasas de pesada en balanza pública, guías de traslado municipales y Documento de Tránsito Electrónico (DTA / SENASA).
*   **Descuento Financiero (Plazo de Pago):** Ajuste por valor tiempo del dinero en ventas con pago a plazo (ej. 30, 45 o 60 días). Permite descontar el flujo futuro a una Tasa Nominal Anual (TNA) de referencia para obtener el **Valor Actual Neto (VAN)** o valor contado equivalente.
*   **Precio Neto de Bolsillo ($/kg):** Rinde financiero efectivo por kilogramo obtenido por el productor tras deducir todas las mermas, gastos comerciales, flete y costo financiero.

---

## 4. Alertas y Señales de Mercado

*   **Variación Diaria ($\Delta\%$):** Variación porcentual del precio promedio entre ruedas consecutivas para una misma categoría canónica.
*   **Oportunidad Estacional:** Desvío positivo de una cotización respecto de la media estacional histórica de la categoría.
*   **Idempotencia de Alerta:** Garantía de no despachar notificaciones duplicadas en un mismo día civil para la misma categoría, tipo de alerta y canal de notificación.

