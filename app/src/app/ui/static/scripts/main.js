// Refactored per SOLID, DRY, KISS, YAGNI

// --- Constants & Utilities ---------------------------------------------
const TIME = {
  INIT_FADE_IN: 200,
  STEP_HIDE: 150,
  STEP_SHOW_DELAY: 200,
  STEP_VISIBLE: 50,
  URL_REMOVE_ANIM: 300
};

const SELECTORS = {
  wizard: '#wizard',
  steps: '.step',
  prevBtn: '#prev-btn',
  nextBtn: '#next-btn',
  submitBtn: '#submit-btn',
  stepItems: '#step-indicator li',
  stepCircles: '#step-indicator .step-circle',
  stepLabels: '#step-indicator .step-label',
  urlsContainer: '#urls-container',
  addUrlBtn: '#add-url',
  form: '#bombard-form',
  respEl: '#response-message',
  fadeIn: '.fade-in',
  transStrategy: '#trans-strategy',
  transformInfo: '#transform-info-text'
};

const TRANSFORM_INFO = {
  JSONATA: 'Use JSONata expressions to transform your data',
  GOTMPL: 'Use Go templates to transform your data',
  JAVASCRIPT: 'Use JavaScript to transform your data',
  DEFAULT: 'Use expressions to transform your data'
};

const ICON = '<i class="fas fa-info-circle mr-1"></i>';

// DOM helpers
const $ = s => document.querySelector(s);
const $$ = s => Array.from(document.querySelectorAll(s));
const getVal = (sel, parser = v => v) => parser($(sel).value.trim());
const checkedVal = name => document.querySelector(`input[name="${name}"]:checked`).value;

// --- Initialization -----------------------------------------------------
document.addEventListener('DOMContentLoaded', init);

function init() {
  initFadeIn();
  initWizard();
  initUrlFields();
  initForm();
  initTransformInfoListener();
}

// --- Fade‑in ------------------------------------------------------------
function initFadeIn() {
  setTimeout(() => $$(SELECTORS.fadeIn).forEach(el => el.classList.add('visible')), TIME.INIT_FADE_IN);
}

// --- Wizard -------------------------------------------------------------
function initWizard() {
  const wizard = $(SELECTORS.wizard);
  if (!wizard) return;
  const steps = $$(SELECTORS.steps);
  let current = 1, total = steps.length;

  const btns = {
    prev: $(SELECTORS.prevBtn),
    next: $(SELECTORS.nextBtn),
    submit: $(SELECTORS.submitBtn)
  };

  const indicators = {
    items: $$(SELECTORS.stepItems),
    circles: $$(SELECTORS.stepCircles),
    labels: $$(SELECTORS.stepLabels)
  };

  const show = step => {
    if (step < 1 || step > total) return;
    steps.forEach(s => {
      s.classList.remove('visible');
      setTimeout(() => s.classList.add('hidden'), TIME.STEP_HIDE);
    });

    const target = wizard.querySelector(`.step[data-step="${step}"]`);
    setTimeout(() => {
      target.classList.remove('hidden');
      setTimeout(() => target.classList.add('visible'), TIME.STEP_VISIBLE);
    }, TIME.STEP_SHOW_DELAY);

    btns.prev.classList.toggle('hidden', step === 1);
    btns.next.classList.toggle('hidden', step === total);
    btns.submit.classList.toggle('hidden', step !== total);

    if (step === total) populateReview();

    indicators.items.forEach((it, i) => it.classList.toggle('step-active', i < step));
    indicators.circles.forEach((c, i) => {
      c.classList.toggle('bg-orange-500', i < step);
      c.classList.toggle('bg-gray-300', i >= step);
    });
    indicators.labels.forEach((l, i) => {
      l.classList.toggle('text-gray-700', i < step);
      l.classList.toggle('text-gray-500', i >= step);
    });

    if (step === 2) updateTransformInfo(getVal(SELECTORS.transStrategy), $(SELECTORS.transformInfo));
  };

  btns.prev.addEventListener('click', () => current > 1 && show(--current));
  btns.next.addEventListener('click', () => current < total && show(++current));
  show(current);
}

// --- URL fields ---------------------------------------------------------
function initUrlFields() {
  const container = $(SELECTORS.urlsContainer);
  const addBtn = $(SELECTORS.addUrlBtn);

  const createField = () => {
    const div = document.createElement('div');
    div.className = 'lb-url flex items-center mb-2 animate__animated animate__fadeIn';
    div.innerHTML = `
      <input type="text" placeholder="https://" class="flex-1 border rounded-l-md py-2 px-3 focus:ring-orange-500">
      <button type="button" class="bg-gray-100 border rounded-r-md px-3 py-2 hover:bg-gray-200 remove-url">
        <i class="fa-solid fa-trash-alt text-gray-600"></i>
      </button>`;
    attachRemove(div.querySelector('.remove-url'), div);
    return div;
  };

  const attachRemove = (btn, parent) => {
    btn.addEventListener('click', () => {
      parent.classList.replace('animate__fadeIn', 'animate__fadeOut');
      setTimeout(() => parent.remove(), TIME.URL_REMOVE_ANIM);
    });
  };

  addBtn.addEventListener('click', () => container.appendChild(createField()));
  $$('.remove-url').forEach(btn => attachRemove(btn, btn.closest('.lb-url')));
  if (!container.querySelector('.lb-url')) container.appendChild(createField());
}

// --- Form & Submission -------------------------------------------------
function initForm() {
  const form = $(SELECTORS.form);
  const resp = $(SELECTORS.respEl);
  form.addEventListener('submit', async e => {
    e.preventDefault();
    if ($(SELECTORS.submitBtn).classList.contains('hidden')) {
      return showResponse('error', 'Please complete all steps before submitting', resp, true);
    }
    const urls = $$('.lb-url input').map(i => i.value.trim()).filter(Boolean);
    if (!urls.length) {
      return showResponse('error', 'Please add at least one target URL', resp);
    }
    showResponse('loading', 'Processing request...', resp);

    const payload = {
      client_context: {
        channel: checkedVal('client_channel'),
        dial_timeout: Number(getVal('#dial-timeout')),
        keep_alive_timeout: Number(getVal('#keepalive-timeout'))
      },
      driver_context: {
        batch_size: Number(getVal('#batch-size')),
        should_store_responses: $('#store-responses').checked
      },
      parser_context: {
        strategy: checkedVal('parser_strategy'),
        file_path: getVal('#file-path')
      },
      load_balancer_context: {
        strategy: checkedVal('lb_strategy'),
        urls
      },
      transformer_context: {
        strategy: getVal('#trans-strategy'),
        method_expression: getVal('#method-expr'),
        endpoint_expression: getVal('#endpoint-expr'),
        headers_expression: getVal('#headers-expr'),
        body_expression: getVal('#body-expr')
      }
    };

    try {
      const res = await fetch('/v1/bombardment', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(payload)
      });
      const data = await res.json();
      if (res.ok) {
        showResponse('success', 'Bombardment started successfully!', resp);
      } else {
        showResponse('error', data.error || 'Error starting bombardment', resp);
      }
    } catch (err) {
      showResponse('error', err.message, resp);
    }
  });
}

// --- Review -------------------------------------------------------------
function populateReview() {
  const summary = $('#review-summary');
  const parser = checkedVal('parser_strategy');
  const file   = getVal('#file-path');
  const trans  = getVal('#trans-strategy');
  const method = getVal('#method-expr');
  const endpoint = getVal('#endpoint-expr');
  const clientChannel = checkedVal('client_channel');
  const dial = getVal('#dial-timeout');
  const keep = getVal('#keepalive-timeout');
  const lbStrat = checkedVal('lb_strategy');
  const urls = $$('.lb-url input').map(i => i.value.trim()).filter(Boolean).join(', ');

  summary.innerHTML = `
    <!-- Source -->
    <div class="flex items-center mb-3 pb-2 border-b">
      <span class="bg-orange-100 text-orange-600 p-1.5 rounded-md mr-2">
        <i class="fas fa-file-import"></i>
      </span>
      <div>
        <h5 class="font-medium">Source</h5>
        <p class="text-gray-600">${parser} (file: ${file})</p>
      </div>
    </div>
    <!-- Transform -->
    <div class="flex items-center mb-3 pb-2 border-b">
      <span class="bg-orange-100 text-orange-600 p-1.5 rounded-md mr-2">
        <i class="fas fa-sliders-h"></i>
      </span>
      <div>
        <h5 class="font-medium">Transform</h5>
        <p class="text-gray-600">Strategy: ${trans}</p>
        <p class="text-gray-600">Method: ${method}, Endpoint: ${endpoint}</p>
        <div class="text-xs mt-1 text-gray-500">
          ${trans === 'custom' ? 'Headers and body expressions configured' : 'Default expressions applied'}
        </div>
      </div>
    </div>
    <!-- Target -->
    <div class="flex items-center">
      <span class="bg-orange-100 text-orange-600 p-1.5 rounded-md mr-2">
        <i class="fas fa-bullseye"></i>
      </span>
      <div>
        <h5 class="font-medium">Target</h5>
        <p class="text-gray-600">
          Channel: ${clientChannel}, Dial: ${dial}ms, KeepAlive: ${keep}ms
        </p>
        <p class="text-gray-600">Load Balancer: ${lbStrat}</p>
        <div class="text-xs mt-1 ${urls ? 'text-gray-500' : 'text-red-500'}">
          ${urls || 'No target URLs configured!'}
        </div>
      </div>
    </div>
  `;
}

// --- Transform Info ----------------------------------------------------
function updateTransformInfo(strategy, el) {
  if (!el) return;
  const text = TRANSFORM_INFO[strategy] || TRANSFORM_INFO.DEFAULT;
  el.innerHTML = `${ICON}<span>${text}</span>`;
}

function initTransformInfoListener() {
  const sel = $(SELECTORS.transStrategy);
  const info = $(SELECTORS.transformInfo);
  sel?.addEventListener('change', () => updateTransformInfo(sel.value, info));
}

// --- Response -----------------------------------------------------------
function showResponse(type, message, container = $(SELECTORS.respEl), resetStep = false) {
  const classes = {
    success: 'p-3 bg-green-100 text-green-600',
    error:   'p-2 bg-red-100 text-red-600',
    loading: 'inline-flex items-center text-gray-600'
  };
  const icon = {
    success: '<i class="fas fa-check-circle mr-2"></i>',
    error:   '<i class="fas fa-exclamation-circle mr-2"></i>',
    loading: '<i class="fas fa-spinner fa-spin mr-2"></i>'
  };
  container.innerHTML = `${icon[type] || ''}${message}`;
  container.className = `mt-4 text-center ${classes[type] || ''}`;
  if (resetStep) {
    const prev = $(SELECTORS.prevBtn);
    prev?.dispatchEvent(new MouseEvent('click'));
  }
}
