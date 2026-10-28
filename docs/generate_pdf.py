import os
import sys

# Inyectar dependencias locales de python si existen
pkgs_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '.python_pkgs'))
if os.path.exists(pkgs_dir) and pkgs_dir not in sys.path:
    sys.path.insert(0, pkgs_dir)

from reportlab.lib.pagesizes import letter
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.platypus import (
    SimpleDocTemplate, Paragraph, Spacer, Table, TableStyle, PageBreak, KeepTogether, Image
)
from reportlab.lib import colors
from reportlab.pdfgen import canvas

# Paleta cromática institucional de Baqueano
PAMPA_GREEN = colors.HexColor('#143224')
PAMPA_LIGHT = colors.HexColor('#2D5A43')
BAQUEANO_GOLD = colors.HexColor('#D49A3D')
RURAL_CREAM = colors.HexColor('#FAF8F5')
MANGA_CARBON = colors.HexColor('#131F18')
MUTED_GREY = colors.HexColor('#64748B')
BORDER_COLOR = colors.HexColor('#E5DFD5')
BG_CARD = colors.HexColor('#F4EFEA')

class NumberedCanvas(canvas.Canvas):
    """Canvas con encabezado y pie de página institucional numerado (Página X de Y)."""
    def __init__(self, *args, **kwargs):
        super(NumberedCanvas, self).__init__(*args, **kwargs)
        self._saved_page_states = []

    def showPage(self):
        self._saved_page_states.append(dict(self.__dict__))
        self._startPage()

    def save(self):
        num_pages = len(self._saved_page_states)
        for state in self._saved_page_states:
            self.__dict__.update(state)
            self.draw_page_decorations(num_pages)
            super(NumberedCanvas, self).showPage()
        super(NumberedCanvas, self).save()

    def draw_page_decorations(self, page_count):
        self.saveState()
        
        # Omitir header en la portada (página 1)
        if self._pageNumber > 1:
            # Header
            self.setFont("Helvetica-Bold", 8)
            self.setFillColor(PAMPA_GREEN)
            self.drawString(54, 755, "BAQUEANO — INTELIGENCIA GANADERA & COTIZACIÓN")
            self.setFont("Helvetica", 8)
            self.setFillColor(MUTED_GREY)
            self.drawRightString(558, 755, "Documentación Técnica y Arquitectura")
            
            # Línea decorativa superior
            self.setStrokeColor(BORDER_COLOR)
            self.setLineWidth(0.75)
            self.line(54, 747, 558, 747)

        # Footer en todas las páginas
        self.setStrokeColor(BORDER_COLOR)
        self.setLineWidth(0.75)
        self.line(54, 45, 558, 45)

        self.setFont("Helvetica", 8)
        self.setFillColor(MUTED_GREY)
        self.drawString(54, 32, "Baqueano v1.0.0 • Monorepo Go • Clean Architecture")
        
        page_str = f"Página {self._pageNumber} de {page_count}"
        self.drawRightString(558, 32, page_str)
        self.restoreState()


def create_documentation_pdf(filename="docs/baqueano_documentacion_tecnica.pdf"):
    os.makedirs(os.path.dirname(filename), exist_ok=True)
    doc = SimpleDocTemplate(
        filename,
        pagesize=letter,
        leftMargin=54,
        rightMargin=54,
        topMargin=54,
        bottomMargin=54
    )

    styles = getSampleStyleSheet()

    # Tipografías y jerarquía visual
    title_style = ParagraphStyle(
        'DocTitle',
        parent=styles['Normal'],
        fontName='Helvetica-Bold',
        fontSize=26,
        leading=32,
        textColor=PAMPA_GREEN,
        spaceAfter=6
    )

    subtitle_style = ParagraphStyle(
        'DocSubtitle',
        parent=styles['Normal'],
        fontName='Helvetica-Bold',
        fontSize=12,
        leading=16,
        textColor=BAQUEANO_GOLD,
        spaceAfter=15
    )

    h1_style = ParagraphStyle(
        'Heading1_Custom',
        parent=styles['Normal'],
        fontName='Helvetica-Bold',
        fontSize=15,
        leading=19,
        textColor=PAMPA_GREEN,
        spaceBefore=16,
        spaceAfter=8,
        keepWithNext=True
    )

    h2_style = ParagraphStyle(
        'Heading2_Custom',
        parent=styles['Normal'],
        fontName='Helvetica-Bold',
        fontSize=11,
        leading=15,
        textColor=PAMPA_LIGHT,
        spaceBefore=10,
        spaceAfter=4,
        keepWithNext=True
    )

    body_style = ParagraphStyle(
        'Body_Custom',
        parent=styles['Normal'],
        fontName='Helvetica',
        fontSize=9.5,
        leading=13.5,
        textColor=MANGA_CARBON,
        spaceAfter=6
    )

    bullet_style = ParagraphStyle(
        'Bullet_Custom',
        parent=styles['Normal'],
        fontName='Helvetica',
        fontSize=9,
        leading=13,
        textColor=MANGA_CARBON,
        leftIndent=15,
        firstLineIndent=-10,
        spaceAfter=4
    )

    callout_style = ParagraphStyle(
        'CalloutText',
        parent=styles['Normal'],
        fontName='Helvetica-Oblique',
        fontSize=9,
        leading=13,
        textColor=PAMPA_GREEN
    )

    meta_style = ParagraphStyle(
        'MetaStyle',
        parent=styles['Normal'],
        fontName='Helvetica',
        fontSize=8.5,
        leading=12,
        textColor=MUTED_GREY
    )

    story = []

    # =========================================================================
    # PORTADA / ENCABEZADO PRINCIPAL
    # =========================================================================
    logo_path = "web/static/img/baqueano_logo_minimalist_star_1793140480833.jpg"
    if os.path.exists(logo_path):
        logo_img = Image(logo_path, width=54, height=54)
        
        header_table = Table(
            [[logo_img, [
                Paragraph("BAQUEANO", title_style),
                Paragraph("INTELIGENCIA GANADERA, COTIZACIÓN EN TIEMPO REAL & LIQUIDACIÓN PECUARIA", subtitle_style)
            ]]],
            colWidths=[65, 439]
        )
        header_table.setStyle(TableStyle([
            ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
            ('LEFTPADDING', (0,0), (-1,-1), 0),
            ('RIGHTPADDING', (0,0), (-1,-1), 0),
            ('BOTTOMPADDING', (0,0), (-1,-1), 0),
            ('TOPPADDING', (0,0), (-1,-1), 0),
        ]))
        story.append(header_table)
    else:
        story.append(Paragraph("BAQUEANO", title_style))
        story.append(Paragraph("INTELIGENCIA GANADERA & LIQUIDACIÓN PECUARIA", subtitle_style))

    story.append(Spacer(1, 10))

    # Cuadro metadata inicial
    meta_data = [
        [Paragraph("<b>Documento:</b> Arquitectura Integral y Especificación Funcional", meta_style),
         Paragraph("<b>Versión:</b> 1.0.0 Production-Ready", meta_style)],
        [Paragraph("<b>Ecosistema:</b> Monorepo Golang 1.22 + Clean Architecture", meta_style),
         Paragraph("<b>Fecha:</b> Octubre 2026", meta_style)]
    ]
    meta_table = Table(meta_data, colWidths=[250, 254])
    meta_table.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,-1), RURAL_CREAM),
        ('BOX', (0,0), (-1,-1), 1, BORDER_COLOR),
        ('INNERGRID', (0,0), (-1,-1), 0.5, BORDER_COLOR),
        ('TOPPADDING', (0,0), (-1,-1), 5),
        ('BOTTOMPADDING', (0,0), (-1,-1), 5),
        ('LEFTPADDING', (0,0), (-1,-1), 8),
        ('RIGHTPADDING', (0,0), (-1,-1), 8),
    ]))
    story.append(meta_table)
    story.append(Spacer(1, 14))

    # =========================================================================
    # SECCIÓN 1: RESUMEN EXPLICATIVO
    # =========================================================================
    story.append(Paragraph("1. Resumen Explicativo", h1_style))
    story.append(Paragraph(
        "<b>Baqueano</b> es una plataforma distribuida de misión crítica diseñada para transformar la comercialización "
        "de hacienda en pie en la República Argentina. Su propósito principal es eliminar la asimetría informativa que "
        "históricamente ha perjudicado al productor ganadero frente a consignatarias, frigoríficos e intermediarios.",
        body_style
    ))
    story.append(Paragraph(
        "La solución combina la <b>extracción automatizada multicanal</b> de precios desde las principales plazas de "
        "referencia (Mercado Agroganadero de Cañuelas — MAG, Rosgan de Rosario, De Campo a Campo y Monasterio Tattersall), "
        "una <b>Capa Anti-Corrupción (ACL)</b> que homogeniza los términos del campo a un modelo ontológico canónico, "
        "un <b>motor reactivo de alertas</b> que detecta fluctuaciones de mercado relevantes (disparadas vía Telegram Bot), "
        "y un <b>simulador financiero de liquidación neta</b> que traduce el precio pactado en remate al verdadero "
        "<b>rinde al bolsillo por kilo entregado</b> al pie del camión.",
        body_style
    ))

    # Callout box
    callout_data = [[
        Paragraph(
            "<b>El concepto de Baqueano:</b> En la cultura rural argentina, el baqueano es quien conoce los caminos, "
            "vados y atajos para llevar la tropa a destino seguro. La plataforma actúa como el guía comercial del productor, "
            "aportando certidumbre numérica ante la volatilidad de plazos, desbastes y fletes.",
            callout_style
        )
    ]]
    callout_table = Table(callout_data, colWidths=[504])
    callout_table.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,-1), BG_CARD),
        ('LEFTPADDING', (0,0), (-1,-1), 12),
        ('RIGHTPADDING', (0,0), (-1,-1), 12),
        ('TOPPADDING', (0,0), (-1,-1), 8),
        ('BOTTOMPADDING', (0,0), (-1,-1), 8),
        ('LINELEFT', (0,0), (-1,-1), 3, BAQUEANO_GOLD),
        ('BOX', (0,0), (-1,-1), 0.5, BORDER_COLOR),
    ]))
    story.append(callout_table)
    story.append(Spacer(1, 14))

    # =========================================================================
    # SECCIÓN 2: EXPLICACIÓN DEL PROBLEMA RESUELTO
    # =========================================================================
    story.append(Paragraph("2. El Problema Resuelto: La 'Ilusión del Precio Bruto'", h1_style))
    story.append(Paragraph(
        "En la comercialización pecuaria tradicional argentina, existe una brecha sustancial entre el precio nominal "
        "fijado en el martillo ($/kg pactado) y el dinero efectivamente acreditado en la cuenta bancaria del productor. "
        "Esta brecha se debe a cuatro distorsiones sistémicas:",
        body_style
    ))

    story.append(Paragraph(
        "• <b>Merma por Desbaste:</b> Pérdida biológica de peso por transporte y orina/bosta. Si no hay balanza en campo, "
        "se pacta un desbaste de entre el 3% y el 8% sobre el peso bruto, reduciendo directamente los kilos facturados.",
        bullet_style
    ))
    story.append(Paragraph(
        "• <b>Asimetría Impositiva en Comisiones:</b> Las consignatarias retienen habitualmente entre un 2% y 3% de comisión. "
        "Mientras la hacienda en pie tributa un IVA del 10.5%, el servicio de intermediación comercial está gravado al <b>21% de IVA</b>, "
        "impactando negativamente en la ecuación de cobro.",
        bullet_style
    ))
    story.append(Paragraph(
        "• <b>Flete e Incidencia Logística:</b> El costo de la jaula de camión se factura como suma fija por kilómetro o viaje, "
        "lo que castiga desproporcionadamente a los lotes más livianos o distantes de los centros de consumo.",
        bullet_style
    ))
    story.append(Paragraph(
        "• <b>Costo Financiero del Plazo Diferido:</b> La hacienda rara vez se liquida de contado. Los plazos habituales son "
        "de 15, 30, 45 o 60 días con cheques de pago diferido. En una economía con tasas anuales (TNA) del 40% al 50%, "
        "esperar 30 o 60 días implica una erosión del capital de trabajo de entre un 3.5% y un 7.5% real.",
        bullet_style
    ))

    story.append(Paragraph(
        "<b>Solución provista por Baqueano:</b> El motor financiero calcula en milisegundos el rinde neto desglosado "
        "y determina con exactitud el <b>Precio Neto por Kilo Neto</b> y el <b>Cobro Neto Efectivo</b>, permitiendo comparar "
        "ofertas disímiles (ej. contado a campo con balanza vs. 45 días en Cañuelas). Adicionalmente, computa la "
        "<b>Relación Ternero/Novillo (Reposición)</b>, informando al ganadero si el mercado favorece la cría o la invernada.",
        body_style
    ))
    story.append(Spacer(1, 14))

    # =========================================================================
    # SECCIÓN 3: DETALLES TÉCNICOS DE LA APLICACIÓN
    # =========================================================================
    story.append(Paragraph("3. Detalles Técnicos de la Aplicación", h1_style))
    story.append(Paragraph(
        "El software fue diseñado bajo los más rigurosos estándares de ingeniería de sistemas distribuidos, "
        "empleando <b>Clean Architecture (Arquitectura Hexagonal)</b> con inversión de dependencias y aislamiento de dominio:",
        body_style
    ))

    tech_table_data = [
        [Paragraph("<b>Capa / Componente</b>", meta_style), 
         Paragraph("<b>Tecnología Seleccionada</b>", meta_style), 
         Paragraph("<b>Justificación de Ingeniería</b>", meta_style)],
        
        [Paragraph("<b>Lenguaje & Runtime</b>", body_style),
         Paragraph("Go 1.22+ (Strict Concurrency)", body_style),
         Paragraph("Rendimiento bare-metal, goroutines ultralivianas, routing HTTP nativo en stdlib y binarios autocontenidos sin dependencias de VM.", body_style)],
        
        [Paragraph("<b>Precisión Matemática</b>", body_style),
         Paragraph("<code>shopspring/decimal</code>", body_style),
         Paragraph("Aritmética decimal de precisión fija. Prohibición de punto flotante IEEE-754 (float64) para evitar errores de redondeo en pesos y dinero.", body_style)],

        [Paragraph("<b>Persistencia Relacional</b>", body_style),
         Paragraph("PostgreSQL 16 + <code>sqlc</code>", body_style),
         Paragraph("Código Go tipado autogenerado directamente desde SQL nativo. Cero overhead de ORM reflectivo; queries validadas en tiempo de compilación.", body_style)],

        [Paragraph("<b>Conexiones de DB</b>", body_style),
         Paragraph("<code>jackc/pgx/v5/pgxpool</code>", body_style),
         Paragraph("Pool de conexiones con keep-alive, timeouts ajustados y soporte para tipos enriquecidos de PostgreSQL.", body_style)],

        [Paragraph("<b>Event Bus & Streams</b>", body_style),
         Paragraph("Redis Streams 7", body_style),
         Paragraph("Canal persistido <code>cotizaciones:stream</code> con grupos de consumo (XREADGROUP/XACK), reintentos y tolerancia a caídas.", body_style)],

        [Paragraph("<b>Alertas & Notificaciones</b>", body_style),
         Paragraph("Telegram Bot API", body_style),
         Paragraph("Despacho asincrónico reactivo con idempotencia diaria estricta en tabla <code>alertas_log</code>.", body_style)],

        [Paragraph("<b>Frontend Web SPA</b>", body_style),
         Paragraph("HTML5 + Tailwind CSS + Chart.js", body_style),
         Paragraph("SPA reactiva servida directamente por el binario de Go. Cero fricción de node_modules; modo offline de respaldo local integrado.", body_style)],
    ]

    t_tech = Table(tech_table_data, colWidths=[110, 140, 254])
    t_tech.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,0), PAMPA_GREEN),
        ('TEXTCOLOR', (0,0), (-1,0), colors.white),
        ('VALIGN', (0,0), (-1,-1), 'TOP'),
        ('GRID', (0,0), (-1,-1), 0.5, BORDER_COLOR),
        ('ROWBACKGROUNDS', (0,1), (-1,-1), [colors.white, RURAL_CREAM]),
        ('TOPPADDING', (0,0), (-1,-1), 5),
        ('BOTTOMPADDING', (0,0), (-1,-1), 5),
        ('LEFTPADDING', (0,0), (-1,-1), 6),
        ('RIGHTPADDING', (0,0), (-1,-1), 6),
    ]))
    story.append(t_tech)
    story.append(Spacer(1, 14))

    # =========================================================================
    # SECCIÓN 4: DETALLES DE EXTRACCIÓN E INGESTA DE DATOS
    # =========================================================================
    story.append(Paragraph("4. Detalles de Extracción e Ingesta de Datos (Scrapers & ACL)", h1_style))
    story.append(Paragraph(
        "La recolección de cotizaciones pecuarias es inherentemente frágil debido a la naturaleza heterogénea de los sitios web "
        "agropecuarios. Baqueano resuelve esta complejidad mediante dos patrones fundamentales:",
        body_style
    ))

    story.append(Paragraph("A. Resiliencia mediante Circuit Breakers y Concurrencia", h2_style))
    story.append(Paragraph(
        "El motor de ingesta (<code>internal/collector/engine.go</code>) lanza goroutines independientes para cada plaza comercial. "
        "Cada fuente está protegida por un <b>Circuit Breaker de 3 estados</b> (Closed, Open, Half-Open). Si una página web externa "
        "se cae, devuelve 500 o sufre un timeout, el circuito se abre evitando saturación de conexiones y garantizando que el resto "
        "de los scrapers finalicen exitosamente sin bloquear la plataforma.",
        body_style
    ))

    story.append(Paragraph("B. Capa Anti-Corrupción (ACL) y Mapeos Ontológicos", h2_style))
    story.append(Paragraph(
        "El desafío del lenguaje agropecuario argentino radica en la dispersión sinonímica (ej. Cañuelas publica 'Vacas Conserva Inferior', "
        "mientras en los remates de feria se denomina 'Vaca China'). El <b>Normalizer ACL</b> realiza el desacoplamiento:",
        body_style
    ))
    story.append(Paragraph(
        "1. <b>Traducción Canónica:</b> Consulta la tabla <code>mapeos_fuente</code> para enlazar el texto crudo a una categoría canónica estandarizada.",
        bullet_style
    ))
    story.append(Paragraph(
        "2. <b>Auto-Registro de Términos Nuevos:</b> Si el scraper extrae una categoría desconocida o un nuevo tramo de peso, el sistema "
        "<b>no se detiene</b>: persiste automáticamente el nuevo término en base de datos con <code>categoria_canon_id = NULL</code> y "
        "<code>verificado_por_humano = false</code>, emitiendo un evento de mapeo pendiente para posterior auditoría humana.",
        bullet_style
    ))
    story.append(Paragraph(
        "3. <b>Normalización Numérica Pampa:</b> Soporte especializado para formatos monetarios y puntuación decimal argentina "
        "(ej. <code>$ 2.450,50</code> es parseado sin pérdida de centavos a <code>2450.50</code>).",
        bullet_style
    ))
    story.append(Spacer(1, 14))

    # =========================================================================
    # SECCIÓN 5: FLUJO GENERAL DE LA APLICACIÓN
    # =========================================================================
    story.append(Paragraph("5. Explicación Detallada del Flujo General", h1_style))
    story.append(Paragraph(
        "El ciclo de vida del dato dentro de Baqueano comprende seis etapas secuenciales y reactivas:",
        body_style
    ))

    flow_steps = [
        ("Paso 1: Scheduler & Extracción",
         "El servicio <code>market-collector</code> despierta según el intervalo programado (ej. cada 30m en horario de mercado con jitter anti-colisión) y dispara los 4 scrapers en paralelo (MAG, Rosgan, De Campo a Campo, Tattersall)."),
        
        ("Paso 2: Normalización & ACL",
         "Las cotizaciones crudas (<code>RawQuote</code>) pasan por el Normalizador ACL, donde se verifican contra el catálogo y se tipan con <code>decimal.Decimal</code>."),
        
        ("Paso 3: Persistencia Idempotente",
         "Las cotizaciones validadas se insertan en PostgreSQL con <code>UpsertCotizacion</code> (resolución de conflictos por fecha + fuente + categoría)."),
        
        ("Paso 4: Publicación en Event Bus",
         "Por cada registro exitoso, se publica un evento <code>cotizaciones:ingeridas</code> en Redis Streams, portando la cotización y su metadata de procedencia."),
        
        ("Paso 5: Evaluación Reactiva de Alertas",
         "El servicio <code>alert-dispatcher</code> consume del stream. Compara el precio contra la cotización del día anterior. Si la variación es relevante (|Δ%| ≥ 4.0%), verifica en <code>alertas_log</code> que no haya sido despachada hoy (idempotencia diaria) y envía un mensaje formateado al bot de Telegram."),
        
        ("Paso 6: Exposición API & Consumo Frontend",
         "El servidor <code>api-server</code> expone las cotizaciones consolidadas, calcula el ratio Ternero/Novillo en tiempo real y provee el endpoint para la calculadora al bolsillo, consumido fluidamente por la interfaz Web SPA."),
    ]

    flow_table_data = []
    for title, desc in flow_steps:
        flow_table_data.append([
            Paragraph(f"<b>{title}</b>", body_style),
            Paragraph(desc, body_style)
        ])

    t_flow = Table(flow_table_data, colWidths=[150, 354])
    t_flow.setStyle(TableStyle([
        ('VALIGN', (0,0), (-1,-1), 'TOP'),
        ('GRID', (0,0), (-1,-1), 0.5, BORDER_COLOR),
        ('ROWBACKGROUNDS', (0,0), (-1,-1), [colors.white, RURAL_CREAM]),
        ('TOPPADDING', (0,0), (-1,-1), 5),
        ('BOTTOMPADDING', (0,0), (-1,-1), 5),
        ('LEFTPADDING', (0,0), (-1,-1), 8),
        ('RIGHTPADDING', (0,0), (-1,-1), 8),
        ('LINELEFT', (0,0), (0,-1), 2.5, PAMPA_GREEN),
    ]))
    story.append(t_flow)
    story.append(Spacer(1, 14))

    # =========================================================================
    # CONCLUSIÓN / HACKATHON READINESS
    # =========================================================================
    story.append(Paragraph("6. Preparación para el Hackathon y Ajuste con Expertos", h1_style))
    story.append(Paragraph(
        "Gracias al desacoplamiento estricto de Clean Architecture, cualquier modificación o sugerencia que aporte "
        "el experto de dominio ganadero durante el hackathon (como ajustar la alícuota de retención de guías municipales, "
        "modificar coeficientes de desbaste por zona geográfica o incorporar nuevas categorías de remates) "
        "se implementa en pocas líneas en la capa de dominio puro (<code>internal/domain/</code>) sin alterar la base de datos "
        "ni los scrapers. La plataforma está 100% lista para demostrar valor comercial y tecnológico inmediato.",
        body_style
    ))

    # Generar el PDF usando el canvas con numeración dinámica
    doc.build(story, canvasmaker=NumberedCanvas)
    print(f"PDF generado con éxito: {filename}")


if __name__ == "__main__":
    out_file = "docs/baqueano_documentacion_tecnica.pdf"
    if len(sys.argv) > 1:
        out_file = sys.argv[1]
    create_documentation_pdf(out_file)
