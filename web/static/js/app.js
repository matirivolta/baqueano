/**
 * Baqueano - Inteligencia Ganadera & Liquidación Neta
 * Lógica Frontend y Reactividad Financiera
 */

const API_BASE = '/api/v1';

// Estado global de la aplicación
const state = {
  cotizaciones: [],
  relacionReposicion: null,
  filtroDestino: 'TODOS',
  categoriaHistoricoId: 1,
  chart: null,
};

// Formateador monetario argentino ($ 1.234,56)
const formatCurrency = (val) => {
  const num = typeof val === 'string' ? parseFloat(val) : val;
  if (isNaN(num)) return '$ 0';
  return '$ ' + num.toLocaleString('es-AR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

// Formateador de números enteros (1.234)
const formatInteger = (val) => {
  const num = typeof val === 'string' ? parseInt(val, 10) : val;
  if (isNaN(num)) return '0';
  return num.toLocaleString('es-AR');
};

// Inicialización de la aplicación
document.addEventListener('DOMContentLoaded', () => {
  initTabs();
  fetchCotizaciones();
  initCalculadora();
  initHistoricoChart();

  // Auto-refresh cada 60 segundos
  setInterval(fetchCotizaciones, 60000);
});

// Manejo de pestañas principales
function initTabs() {
  const tabs = document.querySelectorAll('.nav-tab');
  const sections = {
    'tab-cotizaciones': document.getElementById('sec-cotizaciones'),
    'tab-calculadora': document.getElementById('sec-calculadora'),
    'tab-reposicion': document.getElementById('sec-reposicion'),
  };

  tabs.forEach(tab => {
    tab.addEventListener('click', (e) => {
      e.preventDefault();
      const targetId = tab.dataset.target;

      tabs.forEach(t => {
        t.classList.remove('bg-pampa', 'text-white', 'border-pampa');
        t.classList.add('bg-white', 'text-carbon', 'border-rural');
      });

      tab.classList.remove('bg-white', 'text-carbon', 'border-rural');
      tab.classList.add('bg-pampa', 'text-white', 'border-pampa');

      Object.entries(sections).forEach(([id, el]) => {
        if (!el) return;
        if (id === targetId) {
          el.classList.remove('hidden');
        } else {
          el.classList.add('hidden');
        }
      });
    });
  });
}

// Obtener cotizaciones desde la API
async function fetchCotizaciones() {
  const statusEl = document.getElementById('market-status');
  try {
    const res = await fetch(`${API_BASE}/cotizaciones/resumen`);
    if (!res.ok) throw new Error('Error de conexión a la API');

    const data = await res.json();
    if (!data.cotizaciones || data.cotizaciones.length === 0) {
      console.info('Base de datos sin registros recientes, activando respaldo local.');
      useFallbackData();
      return;
    }

    state.cotizaciones = data.cotizaciones;
    state.relacionReposicion = data.relacion_reposicion || null;

    renderCotizaciones();
    renderReposicion();
    populateCategoriaDropdown();

    if (statusEl) {
      statusEl.innerHTML = `
        <span class="inline-block w-2.5 h-2.5 rounded-full bg-emerald-500 mr-2 online-pulse"></span>
        <span class="text-xs font-semibold text-emerald-800">Mercado en Vivo &bull; ${new Date().toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })}</span>
      `;
    }
  } catch (err) {
    console.warn('API no disponible, usando datos de contingencia:', err);
    if (statusEl) {
      statusEl.innerHTML = `
        <span class="inline-block w-2.5 h-2.5 rounded-full bg-amber-500 mr-2"></span>
        <span class="text-xs font-semibold text-amber-800">Datos Offline &bull; Respaldo Local</span>
      `;
    }
    useFallbackData();
  }
}

// Datos de contingencia para demo local sin conectividad a base de datos
function useFallbackData() {
  state.cotizaciones = [
    // --- 1. FAENA (Mercado Agroganadero de Cañuelas - MAG) ---
    {
      id: 1,
      categoria_codigo: 'NOV_PESADO',
      categoria_nombre: 'Novillos Pesados (+430 kg)',
      fuente: 'MAG',
      tipo_destino: 'FAENA',
      precio_promedio: '2180.00',
      precio_min: '2080.00',
      precio_max: '2250.00',
      volumen_cabezas: 1420,
      variacion_diaria: 1.8,
      unidad: 'kg'
    },
    {
      id: 2,
      categoria_codigo: 'NOVILLITO_MED',
      categoria_nombre: 'Novillitos Medianos (350-390 kg)',
      fuente: 'MAG',
      tipo_destino: 'FAENA',
      precio_promedio: '2260.00',
      precio_min: '2180.00',
      precio_max: '2340.00',
      volumen_cabezas: 2180,
      variacion_diaria: 2.3,
      unidad: 'kg'
    },
    {
      id: 3,
      categoria_codigo: 'NOVILLITO_LIV',
      categoria_nombre: 'Novillitos Livianos (300-350 kg)',
      fuente: 'MAG',
      tipo_destino: 'FAENA',
      precio_promedio: '2310.00',
      precio_min: '2220.00',
      precio_max: '2390.00',
      volumen_cabezas: 1650,
      variacion_diaria: 0.9,
      unidad: 'kg'
    },
    {
      id: 4,
      categoria_codigo: 'VAQUILLONA_ESP',
      categoria_nombre: 'Vaquillonas Especiales (270-340 kg)',
      fuente: 'MAG',
      tipo_destino: 'FAENA',
      precio_promedio: '2240.00',
      precio_min: '2140.00',
      precio_max: '2310.00',
      volumen_cabezas: 1890,
      variacion_diaria: 1.4,
      unidad: 'kg'
    },
    {
      id: 5,
      categoria_codigo: 'VACA_BUENA',
      categoria_nombre: 'Vacas Especiales / Gordas (+420 kg)',
      fuente: 'MAG',
      tipo_destino: 'FAENA',
      precio_promedio: '1680.00',
      precio_min: '1580.00',
      precio_max: '1760.00',
      volumen_cabezas: 1340,
      variacion_diaria: -0.5,
      unidad: 'kg'
    },
    {
      id: 6,
      categoria_codigo: 'VACA_CONSERVA',
      categoria_nombre: 'Vacas Conserva / Manufactura ("China")',
      fuente: 'MAG',
      tipo_destino: 'FAENA',
      precio_promedio: '1420.00',
      precio_min: '1350.00',
      precio_max: '1500.00',
      volumen_cabezas: 1980,
      variacion_diaria: 3.2,
      unidad: 'kg'
    },
    {
      id: 7,
      categoria_codigo: 'TORO',
      categoria_nombre: 'Toros Pesados (+500 kg)',
      fuente: 'MAG',
      tipo_destino: 'FAENA',
      precio_promedio: '1710.00',
      precio_min: '1600.00',
      precio_max: '1820.00',
      volumen_cabezas: 410,
      variacion_diaria: 0.0,
      unidad: 'kg'
    },
    {
      id: 8,
      categoria_codigo: 'MEJ',
      categoria_nombre: 'Macho Entero Joven (MEJ)',
      fuente: 'MAG',
      tipo_destino: 'FAENA',
      precio_promedio: '2190.00',
      precio_min: '2080.00',
      precio_max: '2270.00',
      volumen_cabezas: 350,
      variacion_diaria: 1.1,
      unidad: 'kg'
    },

    // --- 2. INVERNADA (Rosgan, De Campo a Campo, Tattersall) ---
    {
      id: 9,
      categoria_codigo: 'TERNERO_INV',
      categoria_nombre: 'Terneros Invernada (160-180 kg)',
      fuente: 'ROSGAN',
      tipo_destino: 'INVERNADA',
      precio_promedio: '2680.00',
      precio_min: '2550.00',
      precio_max: '2800.00',
      volumen_cabezas: 3400,
      variacion_diaria: 4.2,
      unidad: 'kg'
    },
    {
      id: 10,
      categoria_codigo: 'TERNERO_LIV',
      categoria_nombre: 'Terneros Livianos (hasta 160 kg)',
      fuente: 'DE CAMPO A CAMPO',
      tipo_destino: 'INVERNADA',
      precio_promedio: '2790.00',
      precio_min: '2650.00',
      precio_max: '2920.00',
      volumen_cabezas: 2100,
      variacion_diaria: 3.8,
      unidad: 'kg'
    },
    {
      id: 11,
      categoria_codigo: 'TERNERA_INV',
      categoria_nombre: 'Terneras Invernada (150-170 kg)',
      fuente: 'ROSGAN',
      tipo_destino: 'INVERNADA',
      precio_promedio: '2520.00',
      precio_min: '2400.00',
      precio_max: '2640.00',
      volumen_cabezas: 2750,
      variacion_diaria: 2.0,
      unidad: 'kg'
    },
    {
      id: 12,
      categoria_codigo: 'TERNERO_A_INV',
      categoria_nombre: 'Terneros/as Mixto (140-180 kg)',
      fuente: 'TATTERSALL',
      tipo_destino: 'INVERNADA',
      precio_promedio: '2590.00',
      precio_min: '2480.00',
      precio_max: '2710.00',
      volumen_cabezas: 1820,
      variacion_diaria: 1.5,
      unidad: 'kg'
    },
    {
      id: 13,
      categoria_codigo: 'NOVILLITO_REC',
      categoria_nombre: 'Novillitos de Recría (200-250 kg)',
      fuente: 'ROSGAN',
      tipo_destino: 'INVERNADA',
      precio_promedio: '2380.00',
      precio_min: '2280.00',
      precio_max: '2490.00',
      volumen_cabezas: 1950,
      variacion_diaria: -0.8,
      unidad: 'kg'
    },
    {
      id: 14,
      categoria_codigo: 'VAQUILLONA_REC',
      categoria_nombre: 'Vaquillonas de Recría (200-260 kg)',
      fuente: 'DE CAMPO A CAMPO',
      tipo_destino: 'INVERNADA',
      precio_promedio: '2320.00',
      precio_min: '2210.00',
      precio_max: '2430.00',
      volumen_cabezas: 1480,
      variacion_diaria: 0.4,
      unidad: 'kg'
    },

    // --- 3. CRÍA / VIENTRES ---
    {
      id: 15,
      categoria_codigo: 'VACA_CRIA',
      categoria_nombre: 'Vacas con Cría al Pie (Lo que pisa)',
      fuente: 'TATTERSALL',
      tipo_destino: 'CRIA',
      precio_promedio: '580000.00',
      precio_min: '520000.00',
      precio_max: '650000.00',
      volumen_cabezas: 620,
      variacion_diaria: 1.2,
      unidad: 'cab'
    },
    {
      id: 16,
      categoria_codigo: 'VACA_PRENADA',
      categoria_nombre: 'Vacas Preñadas (Garantía de preñez)',
      fuente: 'ROSGAN',
      tipo_destino: 'CRIA',
      precio_promedio: '890000.00',
      precio_min: '820000.00',
      precio_max: '980000.00',
      volumen_cabezas: 860,
      variacion_diaria: 2.5,
      unidad: 'cab'
    },
    {
      id: 17,
      categoria_codigo: 'VAQ_PRENADA',
      categoria_nombre: 'Vaquillonas Preñadas Especiales',
      fuente: 'ROSGAN',
      tipo_destino: 'CRIA',
      precio_promedio: '940000.00',
      precio_min: '860000.00',
      precio_max: '1050000.00',
      volumen_cabezas: 480,
      variacion_diaria: 3.0,
      unidad: 'cab'
    }
  ];

  state.relacionReposicion = {
    precio_ternero: '2680.00',
    precio_novillo: '2180.00',
    ratio: '1.229',
    interpretacion: 'Relación en equilibrio histórico promedio (1.15 - 1.25).'
  };

  renderCotizaciones();
  renderReposicion();
  populateCategoriaDropdown();
}

// Renderizado de cotizaciones en pantalla
function renderCotizaciones() {
  const container = document.getElementById('cotizaciones-grid');
  if (!container) return;

  const filtered = state.cotizaciones.filter(c => {
    if (state.filtroDestino === 'TODOS') return true;
    return c.tipo_destino === state.filtroDestino;
  });

  if (filtered.length === 0) {
    container.innerHTML = `
      <div class="col-span-full py-12 text-center text-muted">
        No se encontraron cotizaciones para el filtro seleccionado.
      </div>
    `;
    return;
  }

  container.innerHTML = filtered.map(item => {
    const pProm = parseFloat(item.precio_promedio) || 0;
    const pMin = item.precio_min ? parseFloat(item.precio_min) : null;
    const pMax = item.precio_max ? parseFloat(item.precio_max) : null;
    const unidad = item.unidad || 'kg';
    const unidadLabel = unidad === 'cab' ? '/ cabeza' : '/ kg';

    // Estilos por plaza de mercado
    let badgeClass = 'bg-emerald-50 text-emerald-900 border-emerald-200';
    let fuenteLabel = item.fuente;
    if (item.fuente === 'MAG') {
      badgeClass = 'bg-emerald-100 text-emerald-900 border-emerald-300';
      fuenteLabel = 'Cañuelas (MAG)';
    } else if (item.fuente === 'ROSGAN') {
      badgeClass = 'bg-amber-100 text-amber-900 border-amber-300';
      fuenteLabel = 'Rosgan';
    } else if (item.fuente === 'DE CAMPO A CAMPO') {
      badgeClass = 'bg-teal-100 text-teal-900 border-teal-300';
      fuenteLabel = 'De Campo a Campo';
    } else if (item.fuente === 'TATTERSALL') {
      badgeClass = 'bg-sky-100 text-sky-900 border-sky-300';
      fuenteLabel = 'M. Tattersall';
    }

    // Badge de variación interdiaria
    const varDiaria = typeof item.variacion_diaria === 'number' ? item.variacion_diaria : null;
    let varBadge = '';
    if (varDiaria !== null) {
      if (varDiaria > 0) {
        varBadge = `<span class="inline-flex items-center text-[11px] font-bold text-emerald-800 bg-emerald-50 border border-emerald-200 px-1.5 py-0.5 rounded">▲ +${varDiaria.toFixed(1)}%</span>`;
      } else if (varDiaria < 0) {
        varBadge = `<span class="inline-flex items-center text-[11px] font-bold text-rose-800 bg-rose-50 border border-rose-200 px-1.5 py-0.5 rounded">▼ ${varDiaria.toFixed(1)}%</span>`;
      } else {
        varBadge = `<span class="inline-flex items-center text-[11px] font-medium text-gray-600 bg-gray-50 border border-gray-200 px-1.5 py-0.5 rounded">0.0%</span>`;
      }
    }

    // Cabezas comercializadas
    const volumenStr = item.volumen_cabezas ? `${formatInteger(item.volumen_cabezas)} cab.` : null;

    return `
      <div class="bg-white border border-rural rounded-xl p-5 transition-card shadow-sm flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between mb-3">
            <span class="text-xs font-bold uppercase tracking-wider px-2.5 py-1 rounded-md border ${badgeClass}">
              ${fuenteLabel}
            </span>
            <div class="flex items-center gap-1.5">
              ${varBadge}
              <span class="text-[11px] font-bold text-muted uppercase tracking-tight">
                ${item.tipo_destino}
              </span>
            </div>
          </div>

          <h3 class="text-base font-bold text-pampa leading-snug mb-3">
            ${item.categoria_nombre || item.categoria_codigo}
          </h3>

          <div class="my-4">
            <div class="flex items-baseline justify-between mb-1">
              <span class="text-xs font-medium text-muted">Precio Promedio</span>
              ${volumenStr ? `<span class="text-[11px] font-semibold text-muted tracking-tight">${volumenStr}</span>` : ''}
            </div>
            <div class="text-2xl font-black text-pampa tabular-nums">
              ${formatCurrency(pProm)} <span class="text-xs font-semibold text-muted">${unidadLabel}</span>
            </div>
          </div>

          ${pMin && pMax ? `
            <div class="text-xs text-muted font-medium mb-4 flex items-center justify-between border-t border-rural pt-2.5">
              <span>Rango Plaza:</span>
              <span class="tabular-nums font-semibold text-carbon">${formatCurrency(pMin)} - ${formatCurrency(pMax)}</span>
            </div>
          ` : ''}
        </div>

        <button 
          onclick="usarPrecioEnCalculadora(${pProm}, '${item.categoria_nombre || item.categoria_codigo}', ${item.id})"
          class="w-full mt-2 py-2 px-3 bg-cream hover:bg-pampa hover:text-white text-pampa border border-rural hover:border-pampa rounded-lg text-xs font-bold transition duration-150 flex items-center justify-center gap-1.5"
        >
          <svg class="w-4 h-4 text-gold" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path>
          </svg>
          Calcular Liquidación al Bolsillo
        </button>
      </div>
    `;
  }).join('');
}

// Filtrar cotizaciones por destino (Faena, Invernada, etc.)
window.setFiltroDestino = (destino) => {
  state.filtroDestino = destino;
  const filterBtns = document.querySelectorAll('.filter-btn');
  filterBtns.forEach(btn => {
    if (btn.dataset.filtro === destino) {
      btn.classList.remove('bg-white', 'text-carbon', 'border-rural');
      btn.classList.add('bg-pampa', 'text-white', 'border-pampa');
    } else {
      btn.classList.remove('bg-pampa', 'text-white', 'border-pampa');
      btn.classList.add('bg-white', 'text-carbon', 'border-rural');
    }
  });
  renderCotizaciones();
};

// Poblado dinámico del selector de categorías en la calculadora
function populateCategoriaDropdown() {
  const select = document.getElementById('calc-categoria-select');
  if (!select) return;

  const currentVal = select.value;
  select.innerHTML = '<option value="">-- Seleccionar categoría de hacienda --</option>';

  const grupos = {
    'FAENA': { label: 'Faena (Cañuelas MAG)', items: [] },
    'INVERNADA': { label: 'Invernada (Rosgan / Ferias)', items: [] },
    'CRIA': { label: 'Cría / Vientres', items: [] },
  };

  state.cotizaciones.forEach(c => {
    const dest = c.tipo_destino || 'FAENA';
    if (!grupos[dest]) {
      grupos[dest] = { label: dest, items: [] };
    }
    grupos[dest].items.push(c);
  });

  Object.values(grupos).forEach(g => {
    if (g.items.length === 0) return;
    const optgroup = document.createElement('optgroup');
    optgroup.label = g.label;

    g.items.forEach(item => {
      const opt = document.createElement('option');
      opt.value = item.id || item.categoria_codigo;
      const pProm = parseFloat(item.precio_promedio) || 0;
      const unidad = item.unidad === 'cab' ? '/cab' : '/kg';
      opt.textContent = `${item.categoria_nombre || item.categoria_codigo} (${formatCurrency(pProm)}${unidad})`;
      
      opt.dataset.precio = pProm;
      opt.dataset.nombre = item.categoria_nombre || item.categoria_codigo;
      opt.dataset.fuente = item.fuente === 'MAG' ? 'Cañuelas (MAG)' : item.fuente;
      opt.dataset.unidad = item.unidad || 'kg';
      
      // Estimar peso sugerido según nombre de categoría si existe
      let pesoSugerido = null;
      const matchRange = (item.categoria_nombre || '').match(/\((\d+)\s*[-–]\s*(\d+)\s*kg\)/i);
      const matchPlus = (item.categoria_nombre || '').match(/\+(\d+)\s*kg/i);
      const matchUnder = (item.categoria_nombre || '').match(/hasta\s*(\d+)\s*kg/i);

      if (matchRange) {
        pesoSugerido = Math.round((parseInt(matchRange[1], 10) + parseInt(matchRange[2], 10)) / 2);
      } else if (matchPlus) {
        pesoSugerido = parseInt(matchPlus[1], 10) + 20;
      } else if (matchUnder) {
        pesoSugerido = parseInt(matchUnder[1], 10) - 10;
      }
      if (pesoSugerido) {
        opt.dataset.peso = pesoSugerido;
      }

      optgroup.appendChild(opt);
    });

    select.appendChild(optgroup);
  });

  if (currentVal) {
    select.value = currentVal;
  }
}

// Llevar precio seleccionado directamente a la calculadora
window.usarPrecioEnCalculadora = (precio, nombre, id) => {
  const select = document.getElementById('calc-categoria-select');
  if (select && id) {
    select.value = id;
    select.dispatchEvent(new Event('change'));
  } else {
    const precioInput = document.getElementById('calc-precio');
    if (precioInput) {
      precioInput.value = precio;
    }
    const catLabel = document.getElementById('calc-cat-label');
    if (catLabel) {
      catLabel.textContent = `Referencia: ${nombre}`;
      catLabel.classList.remove('hidden');
    }
    recalcularLiquidacion();
  }

  // Cambiar a la pestaña de calculadora
  const calcTab = document.querySelector('[data-target="tab-calculadora"]');
  if (calcTab) calcTab.click();
};

// Inicializar la calculadora interactiva
function initCalculadora() {
  const inputs = [
    'calc-cabezas',
    'calc-peso',
    'calc-precio',
    'calc-desbaste',
    'calc-plazo',
    'calc-tna',
    'calc-comision',
    'calc-flete',
  ];

  inputs.forEach(id => {
    const el = document.getElementById(id);
    if (!el) return;
    el.addEventListener('input', recalcularLiquidacion);
  });

  // Selector dinámico de categoría
  const catSelect = document.getElementById('calc-categoria-select');
  if (catSelect) {
    catSelect.addEventListener('change', () => {
      const selected = catSelect.options[catSelect.selectedIndex];
      if (!selected || !selected.value) {
        const refInfo = document.getElementById('calc-ref-info');
        if (refInfo) refInfo.classList.add('hidden');
        return;
      }

      const precio = parseFloat(selected.dataset.precio);
      const peso = parseFloat(selected.dataset.peso);
      const fuente = selected.dataset.fuente;
      const unidad = selected.dataset.unidad || 'kg';

      const precioInput = document.getElementById('calc-precio');
      if (precioInput && !isNaN(precio) && precio > 0) {
        precioInput.value = precio;
      }

      const pesoInput = document.getElementById('calc-peso');
      if (pesoInput && !isNaN(peso) && peso > 0) {
        pesoInput.value = peso;
      }

      const refInfo = document.getElementById('calc-ref-info');
      if (refInfo) {
        refInfo.textContent = `${fuente} • ${formatCurrency(precio)}${unidad === 'cab' ? '/cab' : '/kg'}`;
        refInfo.classList.remove('hidden');
      }

      recalcularLiquidacion();
    });
  }

  // Sliders sincrónicos con display
  const desbasteSlider = document.getElementById('calc-desbaste');
  const desbasteVal = document.getElementById('calc-desbaste-val');
  if (desbasteSlider && desbasteVal) {
    desbasteSlider.addEventListener('input', (e) => {
      desbasteVal.textContent = `${parseFloat(e.target.value).toFixed(1)}%`;
    });
  }

  populateCategoriaDropdown();
  recalcularLiquidacion();
}

// Recalcular liquidación en tiempo real (instantáneo en cliente con precisión)
function recalcularLiquidacion() {
  const cabezas = parseInt(document.getElementById('calc-cabezas')?.value || '60', 10);
  const pesoBrutoIndiv = parseFloat(document.getElementById('calc-peso')?.value || '390');
  const precioPorKg = parseFloat(document.getElementById('calc-precio')?.value || '2180');
  const pctDesbaste = parseFloat(document.getElementById('calc-desbaste')?.value || '3');
  const plazoDias = parseInt(document.getElementById('calc-plazo')?.value || '30', 10);
  const tna = parseFloat(document.getElementById('calc-tna')?.value || '45');
  const pctComision = parseFloat(document.getElementById('calc-comision')?.value || '2.0');
  const costoFlete = parseFloat(document.getElementById('calc-flete')?.value || '250000');

  if (cabezas <= 0 || pesoBrutoIndiv <= 0 || precioPorKg <= 0) {
    return;
  }

  // 1. Pesajes y desbaste
  const kilosBrutos = cabezas * pesoBrutoIndiv;
  const factorDesbaste = pctDesbaste / 100.0;
  const kilosDesbaste = kilosBrutos * factorDesbaste;
  const kilosNetos = kilosBrutos - kilosDesbaste;

  // 2. Importe Bruto de la Operación
  const importeBruto = kilosNetos * precioPorKg;

  // 3. Deducciones Comerciales
  const montoComision = importeBruto * (pctComision / 100.0);
  const montoIVAComision = montoComision * 0.21; // IVA 21% sobre servicios de consignataria
  const totalDeducciones = montoComision + montoIVAComision + costoFlete;

  // 4. Subtotal antes del costo financiero
  const subtotal = importeBruto - totalDeducciones;

  // 5. Costo Financiero por plazo diferido (TNA)
  const costoFinanciero = subtotal * (tna / 100.0) * (plazoDias / 365.0);

  // 6. Cobro Neto Efectivo al Bolsillo
  const cobroNeto = subtotal - costoFinanciero;

  // 7. Precios netos por kilo
  const precioNetoKiloNeto = kilosNetos > 0 ? cobroNeto / kilosNetos : 0;
  const precioNetoKiloBruto = kilosBrutos > 0 ? cobroNeto / kilosBrutos : 0;

  // Actualizar DOM
  setText('res-kilos-brutos', `${formatInteger(Math.round(kilosBrutos))} kg`);
  setText('res-kilos-desbaste', `-${formatInteger(Math.round(kilosDesbaste))} kg (${pctDesbaste}%)`);
  setText('res-kilos-netos', `${formatInteger(Math.round(kilosNetos))} kg`);
  setText('res-importe-bruto', formatCurrency(importeBruto));

  setText('res-comision', formatCurrency(montoComision));
  setText('res-iva-comision', formatCurrency(montoIVAComision));
  setText('res-flete', formatCurrency(costoFlete));
  setText('res-total-deducciones', `-${formatCurrency(totalDeducciones)}`);

  setText('res-plazo-desc', `${plazoDias} días (TNA ${tna}%)`);
  setText('res-costo-financiero', `-${formatCurrency(costoFinanciero)}`);

  // Grandes cifras destacadas
  setText('res-cobro-neto', formatCurrency(cobroNeto));
  setText('res-precio-neto-kilo', formatCurrency(precioNetoKiloNeto));
  setText('res-precio-neto-bruto', formatCurrency(precioNetoKiloBruto));
}

function setText(id, text) {
  const el = document.getElementById(id);
  if (el) el.textContent = text;
}

// Renderizar sección de Relación Ternero / Novillo (Reposición)
function renderReposicion() {
  if (!state.relacionReposicion) return;

  const ratio = parseFloat(state.relacionReposicion.ratio);
  const ratioEl = document.getElementById('repo-ratio');
  const interEl = document.getElementById('repo-interpretacion');
  const pTerneroEl = document.getElementById('repo-precio-ternero');
  const pNovilloEl = document.getElementById('repo-precio-novillo');

  if (ratioEl) ratioEl.textContent = ratio.toFixed(3);
  if (interEl) interEl.textContent = state.relacionReposicion.interpretacion;
  if (pTerneroEl) pTerneroEl.textContent = formatCurrency(state.relacionReposicion.precio_ternero);
  if (pNovilloEl) pNovilloEl.textContent = formatCurrency(state.relacionReposicion.precio_novillo);

  // Actualizar semáforo visual
  const indicator = document.getElementById('repo-indicator');
  if (indicator) {
    if (ratio > 1.25) {
      indicator.className = 'p-4 rounded-xl bg-amber-50 border border-amber-200 text-amber-900';
    } else if (ratio < 1.15) {
      indicator.className = 'p-4 rounded-xl bg-emerald-50 border border-emerald-200 text-emerald-900';
    } else {
      indicator.className = 'p-4 rounded-xl bg-blue-50 border border-blue-200 text-blue-900';
    }
  }
}

// Inicializar gráfico histórico con Chart.js
function initHistoricoChart() {
  const ctx = document.getElementById('historicoChart');
  if (!ctx || typeof Chart === 'undefined') return;

  const labels = ['Día -25', 'Día -20', 'Día -15', 'Día -10', 'Día -5', 'Hoy'];
  const novilloData = [2050, 2090, 2120, 2100, 2150, 2180];
  const terneroData = [2500, 2560, 2600, 2580, 2640, 2680];

  state.chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: labels,
      datasets: [
        {
          label: 'Ternero Invernada (Rosgan)',
          data: terneroData,
          borderColor: '#D49A3D', // Oro Baqueano
          backgroundColor: 'rgba(212, 154, 61, 0.1)',
          fill: true,
          tension: 0.3,
          borderWidth: 2.5,
          pointRadius: 4,
          pointBackgroundColor: '#D49A3D'
        },
        {
          label: 'Novillo Consumo (Cañuelas MAG)',
          data: novilloData,
          borderColor: '#143224', // Verde Pampa
          backgroundColor: 'transparent',
          tension: 0.3,
          borderWidth: 2.5,
          pointRadius: 4,
          pointBackgroundColor: '#143224'
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'top',
          labels: {
            font: { family: 'Montserrat', size: 12, weight: '600' },
            boxWidth: 14,
            usePointStyle: true,
          }
        },
        tooltip: {
          backgroundColor: '#131F18',
          titleFont: { family: 'Montserrat', size: 12, weight: 'bold' },
          bodyFont: { family: 'Montserrat', size: 12 },
          callbacks: {
            label: (context) => `${context.dataset.label}: $ ${context.parsed.y.toLocaleString('es-AR')} / kg`
          }
        }
      },
      scales: {
        y: {
          grid: { color: '#E5DFD5' },
          ticks: {
            font: { family: 'Montserrat', size: 11 },
            callback: (val) => `$ ${val}`
          }
        },
        x: {
          grid: { display: false },
          ticks: { font: { family: 'Montserrat', size: 11 } }
        }
      }
    }
  });
}

