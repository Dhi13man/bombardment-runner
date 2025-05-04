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
  
  // Cache step indicator elements
  const stepIndicatorItems = document.querySelectorAll('#step-indicator li');
  const stepCircles = Array.from(document.querySelectorAll('#step-indicator .step-circle'));
  const stepLabels = Array.from(document.querySelectorAll('#step-indicator .step-label'));

  function showStep(step) {
    // Hide current step with animation
    steps.forEach(s => {
      s.classList.remove('visible');
      setTimeout(() => {
        s.classList.add('hidden');
      }, 150);
    });
    
    // Show new step with animation
    const newStep = wizard.querySelector(`.step[data-step="${step}"]`);
    setTimeout(() => {
      newStep.classList.remove('hidden');
      setTimeout(() => {
        newStep.classList.add('visible');
      }, 50);
    }, 200);
    
    // Update buttons
    prevBtn.hidden = step === 1;
    nextBtn.hidden = step === totalSteps;
    submitBtn.hidden = step !== totalSteps;
    
    // Populate review if last step
    if (step === totalSteps) populateReview();
    
    // Update step indicator styling
    stepIndicatorItems.forEach((item, idx) => {
      // Remove active class from all steps
      item.classList.remove('step-active');
      
      // Add active class to current and previous steps
      if (idx < step) {
        item.classList.add('step-active');
      }
    });
    
    // Update circles and connecting lines
    stepCircles.forEach((circle, idx) => {
      if (idx < step) {
        circle.classList.remove('bg-gray-300');
        circle.classList.add('bg-orange-500');
      } else {
        circle.classList.remove('bg-orange-500');
        circle.classList.add('bg-gray-300');
      }
    });
    
    // Update labels
    stepLabels.forEach((label, idx) => {
      if (idx < step) {
        label.classList.remove('text-gray-500');
        label.classList.add('text-gray-700');
      } else {
        label.classList.remove('text-gray-700');
        label.classList.add('text-gray-500');
      }
    });
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
      <div class="flex items-center mb-3 pb-2 border-b border-gray-200">
        <span class="bg-orange-100 text-orange-600 p-1.5 rounded-md mr-2">
          <i class="fas fa-file-import"></i>
        </span>
        <div>
          <h5 class="font-medium">Source</h5>
          <p class="text-gray-600">${parserStrategy} (file: ${filePath})</p>
        </div>
      </div>
      
      <div class="flex items-center mb-3 pb-2 border-b border-gray-200">
        <span class="bg-orange-100 text-orange-600 p-1.5 rounded-md mr-2">
          <i class="fas fa-sliders-h"></i>
        </span>
        <div>
          <h5 class="font-medium">Transform</h5>
          <p class="text-gray-600">Strategy: ${transStrategy}</p>
          <p class="text-gray-600">Method: ${methodExpr}, Endpoint: ${endpointExpr}</p>
          <div class="text-xs mt-1 text-gray-500">Headers and body expressions configured</div>
        </div>
      </div>
      
      <div class="flex items-center">
        <span class="bg-orange-100 text-orange-600 p-1.5 rounded-md mr-2">
          <i class="fas fa-bullseye"></i>
        </span>
        <div>
          <h5 class="font-medium">Target</h5>
          <p class="text-gray-600">Channel: ${clientChannel}, Dial: ${dialTimeout}ms, KeepAlive: ${keepAlive}ms</p>
          <p class="text-gray-600">Load Balancer: ${lbStrategy}</p>
          <div class="text-xs mt-1 ${urls ? 'text-gray-500' : 'text-red-500'}">${urls ? `URLs: ${urls}` : 'No target URLs configured!'}</div>
        </div>
      </div>
    `;
  }

  prevBtn.addEventListener('click', () => { 
    if (currentStep > 1) {
      currentStep--;
      showStep(currentStep);
    }
  });
  
  nextBtn.addEventListener('click', () => { 
    if (currentStep < totalSteps) {
      currentStep++;
      showStep(currentStep);
    }
  });
  
  showStep(currentStep);

  // URL add/remove setup
  const addUrlBtn = document.getElementById('add-url');
  const urlsContainer = document.getElementById('urls-container');

  function createUrlField() {
    const div = document.createElement('div');
    div.className = 'lb-url flex items-center mb-2 animate__animated animate__fadeIn';
    div.innerHTML = `
      <input type="text" placeholder="https://" title="Target URL" class="flex-1 border border-gray-300 rounded-l-md py-2 px-3 focus:ring-orange-500 focus:border-orange-500">
      <button type="button" aria-label="Remove URL" class="bg-gray-100 border border-gray-300 border-l-0 rounded-r-md px-3 py-2 hover:bg-gray-200 remove-url">
        <i class="fa-solid fa-trash-alt text-gray-600"></i>
      </button>
    `;
    const removeBtn = div.querySelector('.remove-url');
    removeBtn.addEventListener('click', () => {
      div.classList.add('animate__fadeOut');
      setTimeout(() => div.remove(), 300);
    });
    return div;
  }

  addUrlBtn.addEventListener('click', () => urlsContainer.appendChild(createUrlField()));
  urlsContainer.querySelectorAll('.remove-url').forEach(btn => btn.addEventListener('click', () => {
    const parent = btn.closest('.lb-url');
    if (parent) {
      parent.classList.add('animate__fadeOut');
      setTimeout(() => parent.remove(), 300);
    }
  }));

  // Form submission
  const form = document.getElementById('bombard-form');
  const respEl = document.getElementById('response-message');

  form.addEventListener('submit', async e => {
    e.preventDefault();
    
    // Prevent submission until last step is reached
    if (currentStep !== totalSteps) {
      respEl.textContent = 'Please complete all steps before submitting.';
      respEl.className = 'text-red-600';
      return;
    }
    
    // Validate required fields
    const urls = Array.from(document.querySelectorAll('.lb-url input'))
                  .map(i => i.value.trim())
                  .filter(u => u);
                  
    if (urls.length === 0) {
      respEl.innerHTML = '<div class="p-2 bg-red-100 text-red-600 rounded-md inline-flex items-center"><i class="fas fa-exclamation-triangle mr-2"></i> Please add at least one target URL</div>';
      respEl.className = 'mt-4 text-center';
      return;
    }
    
    respEl.innerHTML = '<div class="inline-flex items-center"><i class="fas fa-spinner fa-spin mr-2"></i> Processing request...</div>';
    respEl.className = 'mt-4 text-center text-gray-600';
    
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
        urls: urls
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
      const res = await fetch('/v1/bombardment', { 
        method: 'POST', 
        headers: { 'Content-Type': 'application/json' }, 
        body: JSON.stringify(payload) 
      });
      
      const data = await res.json();
      
      if (res.ok) { 
        respEl.innerHTML = '<div class="p-3 bg-green-100 text-green-600 rounded-md inline-flex items-center"><i class="fas fa-check-circle mr-2"></i> Bombardment started successfully!</div>';
        respEl.className = 'mt-4 text-center'; 
      } else { 
        respEl.innerHTML = `<div class="p-3 bg-red-100 text-red-600 rounded-md inline-flex items-center"><i class="fas fa-exclamation-triangle mr-2"></i> ${data.error || 'Error starting bombardment'}</div>`;
        respEl.className = 'mt-4 text-center'; 
      }
    } catch (err) {
      respEl.innerHTML = `<div class="p-3 bg-red-100 text-red-600 rounded-md inline-flex items-center"><i class="fas fa-exclamation-triangle mr-2"></i> ${err.message}</div>`;
      respEl.className = 'mt-4 text-center';
    }
  });
  
  // Initialize first URL field if none exist
  if (urlsContainer.querySelectorAll('.lb-url').length === 0) {
    urlsContainer.appendChild(createUrlField());
  }
});
