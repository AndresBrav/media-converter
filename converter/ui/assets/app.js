document.addEventListener('DOMContentLoaded', () => {

  /* ─── DOM refs ───────────────────────────────── */
  const inputDir       = document.getElementById('input-dir');
  const outputDir      = document.getElementById('output-dir');
  const formatSelect   = document.getElementById('format-select');
  const workersInput   = document.getElementById('workers-input');
  const qualitySlider  = document.getElementById('quality-slider');
  const qualityBadge   = document.getElementById('quality-badge');
  const qualityField   = document.getElementById('quality-field');
  const widthInput     = document.getElementById('width-input');
  const heightInput    = document.getElementById('height-input');
  const watermarkPath  = document.getElementById('watermark-path');
  const recursiveToggle= document.getElementById('recursive-toggle');
  const thumbnailToggle= document.getElementById('thumbnail-toggle');

  const btnBrowseInput   = document.getElementById('btn-browse-input');
  const btnBrowseOutput  = document.getElementById('btn-browse-output');
  const btnBrowseWmark   = document.getElementById('btn-browse-watermark');
  const btnStart         = document.getElementById('btn-start');
  const btnStop          = document.getElementById('btn-stop');
  const btnClearLogs     = document.getElementById('btn-clear-logs');

  const progressFill  = document.getElementById('progress-fill');
  const progressLabel = document.getElementById('progress-label');
  const statRatio     = document.getElementById('stat-ratio');
  const statErrors    = document.getElementById('stat-errors');
  const statTime      = document.getElementById('stat-time');

  const metricCompleted  = document.getElementById('metric-completed');
  const metricFailed     = document.getElementById('metric-failed');
  const metricTime       = document.getElementById('metric-time');
  const metricWorkers    = document.getElementById('metric-workers-count');

  const workersGrid = document.getElementById('workers-grid');
  const terminal    = document.getElementById('terminal-logs');
  const autoScroll  = document.getElementById('auto-scroll');
  const serverDot   = document.getElementById('server-dot');
  const serverLabel = document.getElementById('server-label');
  const portDisplay = document.getElementById('port-display');

  /* ─── Tabs ───────────────────────────────────── */
  document.querySelectorAll('.nav-item').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.nav-item').forEach(b => b.classList.remove('active'));
      document.querySelectorAll('.tab-pane').forEach(p => p.classList.remove('active'));
      btn.classList.add('active');
      const tab = btn.dataset.tab;
      document.getElementById(`tab-${tab}`).classList.add('active');
    });
  });

  /* ─── Port display ───────────────────────────── */
  const port = window.location.port || '8080';
  if (portDisplay) portDisplay.textContent = port;

  /* ─── Default workers = CPU cores ───────────── */
  if (navigator.hardwareConcurrency) {
    workersInput.value = navigator.hardwareConcurrency;
    workersInput.max   = navigator.hardwareConcurrency * 2;
  }

  /* ─── Quality slider ─────────────────────────── */
  qualitySlider.addEventListener('input', () => {
    qualityBadge.textContent = `${qualitySlider.value}%`;
  });

  formatSelect.addEventListener('change', () => {
    const isPng = formatSelect.value === '.png';
    qualityField.style.opacity = isPng ? '.4' : '1';
    qualitySlider.disabled     = isPng;
    qualityBadge.textContent   = isPng ? 'N/A' : `${qualitySlider.value}%`;
  });

  /* ─── Clear logs ─────────────────────────────── */
  btnClearLogs.addEventListener('click', () => {
    terminal.innerHTML = '';
    log('Logs limpiados.', 'system');
  });

  /* ─── Logger ─────────────────────────────────── */
  function log(text, type = 'system') {
    const line = document.createElement('div');
    line.className = `log-line log-${type}`;
    const t = new Date().toLocaleTimeString('es', { hour12: false });
    line.textContent = `[${t}] ${text}`;
    terminal.appendChild(line);
    if (autoScroll.checked) terminal.scrollTop = terminal.scrollHeight;
  }

  /* ─── Browse dialogs ─────────────────────────── */
  async function browse(field, type) {
    try {
      const res  = await fetch(`/api/browse?type=${type}`);
      const data = await res.json();
      if (res.ok && data.path) {
        field.value = data.path;
        log(`Seleccionado: ${data.path}`, 'info');
      } else {
        log(`Selección cancelada.`, 'system');
      }
    } catch (e) {
      log(`Error al abrir diálogo: ${e.message}`, 'error');
    }
  }
  btnBrowseInput.addEventListener ('click', () => browse(inputDir, 'directory'));
  btnBrowseOutput.addEventListener('click', () => browse(outputDir, 'directory'));
  btnBrowseWmark.addEventListener ('click', () => browse(watermarkPath, 'file'));

  /* ─── Workers grid ───────────────────────────── */
  function buildWorkersGrid(n) {
    workersGrid.innerHTML = '';
    for (let i = 1; i <= n; i++) {
      const pill = document.createElement('div');
      pill.className = 'worker-pill';
      pill.id = `worker-${i}`;
      pill.innerHTML = `
        <div class="worker-pill-header">
          <span>W-${i}</span>
          <span class="worker-dot"></span>
        </div>
        <div class="worker-content"><span class="worker-idle">inactivo</span></div>`;
      workersGrid.appendChild(pill);
    }
    metricWorkers.textContent = n;
  }

  function setWorkerBusy(id, file) {
    const pill = document.getElementById(`worker-${id}`);
    if (!pill) return;
    pill.classList.add('busy');
    pill.querySelector('.worker-content').innerHTML =
      `<span class="worker-filename" title="${file}">${file.split(/[/\\]/).pop()}</span>`;
  }

  function setWorkerIdle(id) {
    const pill = document.getElementById(`worker-${id}`);
    if (!pill) return;
    pill.classList.remove('busy');
    pill.querySelector('.worker-content').innerHTML = `<span class="worker-idle">inactivo</span>`;
  }

  /* ─── Timer (client-side fallback display) ───── */
  let timerInterval = null;
  let timerStart    = null;

  function startTimer() {
    timerStart = Date.now();
    clearInterval(timerInterval);
    timerInterval = setInterval(() => {
      const s = ((Date.now() - timerStart) / 1000).toFixed(1);
      statTime.textContent    = `${s}s`;
      metricTime.textContent  = `${s}s`;
    }, 100);
  }

  function stopTimer(finalSecs) {
    clearInterval(timerInterval);
    timerInterval = null;
    if (finalSecs !== undefined) {
      const display = `${finalSecs.toFixed(1)}s`;
      statTime.textContent   = display;
      metricTime.textContent = display;
    }
  }

  /* ─── UI state helpers ───────────────────────── */
  function setRunning(running) {
    btnStart.disabled = running;
    btnStop.disabled  = !running;

    document.querySelectorAll(
      '#converter-form input, #converter-form select, #converter-form button.btn-ghost'
    ).forEach(el => el.disabled = running);

    serverDot.className   = running ? 'status-dot processing' : 'status-dot';
    serverLabel.textContent = running ? 'Procesando…' : 'Conectado';

    if (!running) {
      // re-enable quality if not png
      if (formatSelect.value !== '.png') qualitySlider.disabled = false;
    }
  }

  function resetProgress() {
    progressFill.style.width = '0%';
    progressLabel.textContent = 'Sin proceso activo';
    statRatio.textContent = '0 / 0';
    statErrors.textContent = '0 errores';
    statErrors.classList.remove('ok');
    statTime.textContent = '0.0s';
    metricCompleted.textContent = '0';
    metricFailed.textContent    = '0';
    metricTime.textContent      = '0.0s';
  }

  /* ─── SSE ────────────────────────────────────── */
  let evSource = null;

  function connectSSE() {
    if (evSource) { evSource.close(); evSource = null; }
    evSource = new EventSource('/api/events');
    evSource.onmessage = e => {
      try { handleEvent(JSON.parse(e.data)); }
      catch (_) {}
    };
    evSource.onerror = () => {
      log('Conexión SSE interrumpida.', 'warn');
    };
  }

  function disconnectSSE() {
    if (evSource) { evSource.close(); evSource = null; }
  }

  function handleEvent(ev) {
    switch (ev.type) {

      case 'start':
        progressLabel.textContent = `Convirtiendo ${ev.total} archivo${ev.total !== 1 ? 's' : ''}…`;
        statRatio.textContent     = `0 / ${ev.total}`;
        startTimer();
        log(`Iniciado: ${ev.total} archivos encontrados.`, 'info');
        break;

      case 'worker_start':
        setWorkerBusy(ev.worker_id, ev.input_path);
        break;

      case 'worker_end': {
        setWorkerIdle(ev.worker_id);

        const pct = ev.total > 0 ? Math.round((ev.current / ev.total) * 100) : 0;
        progressFill.style.width      = `${pct}%`;
        progressLabel.textContent     = `${pct}% completado`;
        statRatio.textContent         = `${ev.current} / ${ev.total}`;
        metricCompleted.textContent   = ev.current - ev.failed;
        metricFailed.textContent      = ev.failed;

        const errCount = ev.failed;
        statErrors.textContent = `${errCount} error${errCount !== 1 ? 'es' : ''}`;
        statErrors.classList.toggle('ok', errCount === 0);

        // Actualizar tiempo desde el backend (CORREGIDO)
        if (ev.elapsed !== undefined) {
          const display = `${ev.elapsed.toFixed(1)}s`;
          statTime.textContent   = display;
          metricTime.textContent = display;
        }

        const fname = ev.input_path.split(/[/\\]/).pop();
        if (ev.status === 'success') {
          const oname = ev.output_path.split(/[/\\]/).pop();
          log(`✓ Worker ${ev.worker_id}  ${fname} → ${oname}`, 'success');
        } else {
          log(`✗ Worker ${ev.worker_id}  ${fname}  ${ev.error}`, 'error');
        }
        break;
      }

      case 'complete':
        stopTimer(ev.elapsed);
        setRunning(false);
        disconnectSSE();

        // All workers idle
        document.querySelectorAll('.worker-pill').forEach((_, i) => setWorkerIdle(i + 1));

        progressFill.style.width    = '100%';
        progressLabel.textContent   = '¡Conversión completada!';
        metricCompleted.textContent = ev.current - ev.failed;
        metricFailed.textContent    = ev.failed;
        statRatio.textContent       = `${ev.current} / ${ev.total}`;

        log(`Completado — ${ev.current - ev.failed} exitosos, ${ev.failed} errores, ${ev.elapsed.toFixed(1)}s`, 'success');
        break;

      case 'stopped':
        stopTimer(ev.elapsed);
        setRunning(false);
        disconnectSSE();
        document.querySelectorAll('.worker-pill').forEach((_, i) => setWorkerIdle(i + 1));
        progressLabel.textContent = 'Proceso cancelado';
        log('Proceso detenido por el usuario.', 'warn');
        break;
    }
  }

  /* ─── Start button ───────────────────────────── */
  btnStart.addEventListener('click', async () => {
    if (!inputDir.value.trim() || !outputDir.value.trim()) {
      alert('Selecciona el directorio de entrada y salida antes de iniciar.');
      return;
    }

    const n = parseInt(workersInput.value) || 4;
    resetProgress();
    buildWorkersGrid(n);
    setRunning(true);
    connectSSE();

    const body = {
      inputDir:  inputDir.value.trim(),
      outputDir: outputDir.value.trim(),
      format:    formatSelect.value,
      workers:   n,
      quality:   parseInt(qualitySlider.value),
      width:     parseInt(widthInput.value)  || 0,
      height:    parseInt(heightInput.value) || 0,
      watermark: watermarkPath.value.trim(),
      thumbnail: thumbnailToggle.checked,
      recursive: recursiveToggle.checked,
    };

    log(`Enviando configuración → ${body.format.toUpperCase()}, ${body.workers} workers`, 'info');

    try {
      const res = await fetch('/api/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Error al iniciar');
      }
    } catch (e) {
      log(`Error: ${e.message}`, 'error');
      setRunning(false);
      disconnectSSE();
      stopTimer();
    }
  });

  /* ─── Stop button ────────────────────────────── */
  btnStop.addEventListener('click', async () => {
    log('Enviando señal de cancelación…', 'warn');
    btnStop.disabled = true;
    try {
      await fetch('/api/stop', { method: 'POST' });
    } catch (e) {
      log(`Error al detener: ${e.message}`, 'error');
    }
  });

});
