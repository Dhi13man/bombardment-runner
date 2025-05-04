// UI logic for Bombardment form (wizard + URL add/remove + submission)

document.addEventListener('DOMContentLoaded', () => {
  // Wizard setup
  const wizard = document.getElementById('wizard');
  const steps = wizard.querySelectorAll('.step');
  let currentStep = 1;
  const totalSteps = steps.length;
  const prevBtn = document.getElementById('prev-btn');
  const nextBtn = document.getElementById('next-btn');
  const submitBtn = document.getElementById('submit-btn');

  function showStep(step) {
    steps.forEach(s => s.classList.add('hidden'));
    wizard.querySelector(`.step[data-step="${step}"]`).classList.remove('hidden');
    prevBtn.hidden = step === 1;
    nextBtn.hidden = step === totalSteps;
    submitBtn.hidden = step !== totalSteps;
    if (step === totalSteps) populateReview();
  }

  function populateReview() {
    const summary = document.getElementById('review-summary');
    const parserStrategy = document.querySelector('input[name="parser_strategy"]:checked').value;
    const filePath = document.getElementById('file-path').value;
    const transStrategy = document.getElementById('trans-strategy').value;
    const methodExpr = document.getElementById('method-expr').value;
    const endpointExpr = document.getElementById('endpoint-expr').value;
    const headersExpr = document.getElementById('headers-expr').value;
    const bodyExpr = document.getElementById('body-expr').value;
    const clientChannel = document.querySelector('input[name="client_channel"]:checked').value;
    const dialTimeout = document.getElementById('dial-timeout').value;
    const keepAlive = document.getElementById('keepalive-timeout').value;
    const lbStrategy = document.querySelector('input[name="lb_strategy"]:checked').value;
    const urls = Array.from(document.querySelectorAll('.lb-url input'))
                    .map(i => i.value.trim())
                    .filter(u => u)
                    .join(', ');

    summary.innerHTML = `
      <p><strong>Source:</strong> ${parserStrategy} (file: ${filePath})</p>
      <p><strong>Transform:</strong> ${transStrategy}</p>
      <p>Method: ${methodExpr}, Endpoint: ${endpointExpr}</p>
      <p>Headers: ${headersExpr}</p>
      <p>Body: ${bodyExpr}</p>
      <p><strong>Target:</strong> Channel: ${clientChannel}, Dial: ${dialTimeout}ms, KeepAlive: ${keepAlive}ms</p>
      <p>Load Balancer: ${lbStrategy} (URLs: ${urls})</p>
    `;
  }

  prevBtn.addEventListener('click', () => { if (currentStep > 1) showStep(--currentStep); });
  nextBtn.addEventListener('click', () => { if (currentStep < totalSteps) showStep(++currentStep); });
  showStep(currentStep);

  // URL add/remove setup
  const addUrlBtn = document.getElementById('add-url');
  const urlsContainer = document.getElementById('urls-container');

  function createUrlField() {
    const div = document.createElement('div');
    div.className = 'lb-url flex items-center';
    div.innerHTML = `
      <input type="text" placeholder="https://" title="Target URL" class="flex-1 border border-gray-300 rounded-l-md py-2 px-3 focus:ring-orange-500 focus:border-orange-500">
      <button type="button" aria-label="Remove URL" class="bg-gray-100 border border-gray-300 border-l-0 rounded-r-md px-3 py-2 hover:bg-gray-200 remove-url">
        <i class="fa-solid fa-trash-alt text-gray-600"></i>
      </button>
    `;
    const removeBtn = div.querySelector('.remove-url');
    removeBtn.addEventListener('click', () => div.remove());
    return div;
  }

  addUrlBtn.addEventListener('click', () => urlsContainer.appendChild(createUrlField()));
  urlsContainer.querySelectorAll('.remove-url').forEach(btn => btn.addEventListener('click', () => {
    const parent = btn.closest('.lb-url'); if (parent) parent.remove();
  }));

  // Form submission
  const form = document.getElementById('bombard-form');
  const respEl = document.getElementById('response-message');

  form.addEventListener('submit', async e => {
    e.preventDefault(); respEl.textContent = '';
    const payload = {
      client_context: {
        channel: document.querySelector('input[name="client_channel"]:checked').value,
        dial_timeout: parseInt(document.getElementById('dial-timeout').value, 10),
        keep_alive_timeout: parseInt(document.getElementById('keepalive-timeout').value, 10)
      },
      driver_context: {
        batch_size: parseInt(document.getElementById('batch-size').value, 10),
        should_store_responses: document.getElementById('store-responses').checked
      },
      parser_context: {
        strategy: document.querySelector('input[name="parser_strategy"]:checked').value,
        file_path: document.getElementById('file-path').value
      },
      load_balancer_context: {
        strategy: document.querySelector('input[name="lb_strategy"]:checked').value,
        urls: Array.from(document.querySelectorAll('.lb-url input')).map(i => i.value.trim()).filter(u => u)
      },
      transformer_context: {
        strategy: document.getElementById('trans-strategy').value,
        method_expression: document.getElementById('method-expr').value,
        endpoint_expression: document.getElementById('endpoint-expr').value,
        headers_expression: document.getElementById('headers-expr').value,
        body_expression: document.getElementById('body-expr').value
      }
    };
    try {
      const res = await fetch('/v1/bombardment', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) });
      const data = await res.json();
      if (res.ok) { respEl.textContent = 'Bombardment started!'; respEl.className = 'text-green-600'; }
      else { respEl.textContent = data.error || 'Error starting bombardment'; respEl.className = 'text-red-600'; }
    } catch (err) {
      respEl.textContent = err.message; respEl.className = 'text-red-600';
    }
  });
});
