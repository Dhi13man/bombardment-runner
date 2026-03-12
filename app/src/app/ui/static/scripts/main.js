// Refactored per SOLID, DRY, KISS, YAGNI

// --- Constants & Utilities ---------------------------------------------
const TIME = {
  INIT_FADE_IN: 200,
  STEP_HIDE: 150,
  STEP_SHOW_DELAY: 200,
  STEP_VISIBLE: 50,
  URL_REMOVE_ANIM: 300,
  VALIDATION_DEBOUNCE: 300
};

const SELECTORS = {
  wizard: '#wizard',
  steps: '.step',
  prevBtn: '#prev-btn',
  nextBtn: '#next-btn',
  submitBtn: '#submit-btn',
  stepItems: '#step-indicator .step-item',
  stepArrows: '#step-indicator .step-arrow',
  stepLabels: '#step-indicator .step-label',
  urlsContainer: '#urls-container',
  addUrlBtn: '#add-url',
  form: '#bombard-form',
  respEl: '#response-message',
  fadeIn: '.fade-in',
  transStrategy: '#trans-strategy',
  transformInfo: '#transform-info-text',
  fileInput: '#file-input',
  filePathDisplay: '#file-path-display',
  filePathHidden: '#file-path',
  fileDetails: '#file-details',
  fileSize: '#file-size',
  fileModified: '#file-modified',
  filePathNote: '#file-path-note',
  fileContentB64: '#file-content-b64',
  storeResponses: '#store-responses',
  responsesPathContainer: '#responses-path-container'
};

const TRANSFORM_INFO = {
  JSONATA: 'Use JSONata expressions to transform your data',
  GOTMPL: 'Use Go templates to transform your data',
  JAVASCRIPT: 'Use JavaScript to transform your data',
  DEFAULT: 'Use expressions to transform your data'
};

const VALIDATION = {
  FILE_PATH_REGEX: /^(\.[\/\\])?([a-zA-Z0-9_\-\/\\]+)\.([a-zA-Z0-9]+)$/,
  URL_REGEX: /^https?:\/\/([a-zA-Z0-9][-a-zA-Z0-9]*(\.[a-zA-Z0-9][-a-zA-Z0-9]*)+|localhost)(:[0-9]{1,5})?(\/[-a-zA-Z0-9()@:%_\+.~#?&//=]*)?$/,
  MIN_BATCH_SIZE: 1,
  MAX_BATCH_SIZE: 10000,
  MIN_TIMEOUT: 100,
  MAX_TIMEOUT: 60000
};

const ICON = {
  INFO: '<i class="fas fa-info-circle mr-1"></i>',
  ERROR: '<i class="fas fa-exclamation-circle mr-1"></i>',
  SUCCESS: '<i class="fas fa-check-circle mr-1"></i>'
};

// DOM helpers
const $ = s => document.querySelector(s);
const $$ = s => Array.from(document.querySelectorAll(s));
const getVal = (sel, parser = v => v) => parser($(sel)?.value?.trim() || '');
const checkedVal = name => document.querySelector(`input[name="${name}"]:checked`)?.value;

// Security: HTML escaping to prevent XSS
function escapeHtml(str) {
  if (str == null) return '';
  const div = document.createElement('div');
  div.appendChild(document.createTextNode(String(str)));
  return div.innerHTML;
}

// Validation state
const validationState = {
  step1: false,
  step2: false,
  step3: false,
  step4: false,
  
  // Detailed validation states for review
  source: {
    valid: false,
    errors: []
  },
  transform: {
    valid: false,
    errors: []
  },
  target: {
    valid: false,
    errors: []
  },
  driver: {
    valid: false,
    errors: []
  },
  
  // Helper to get overall status
  get isValid() {
    return this.step1 && this.step2 && this.step3 && this.step4;
  }
};

// Debounce function for validation
function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

// --- Initialization -----------------------------------------------------
document.addEventListener('DOMContentLoaded', init);

function init() {
  initFadeIn();
  initWizard();
  initUrlFields();
  initForm();
  initTransformInfoListener();
  initValidation();
  initConfigurationIssues();
  initFileInput();
  initParserStrategyListeners();
  initStoreResponsesListener();
  initJobProgressActions();
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
    arrows: $$(SELECTORS.stepArrows),
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

    if (step === total) {
      populateReview(); 
      // Initialize issues section when reaching review step
      const issuesSection = $('#issues-section');
      if (issuesSection) {
        setTimeout(() => {
          displayConfigurationIssues();
          // Show issues section if there are issues
          const issueCount = $('#issue-count');
          if (issueCount && parseInt(issueCount.textContent) > 0) {
            issuesSection.classList.remove('hidden');
          }
        }, 300); // Slight delay to ensure the review summary is populated
      }
    }

    // Update the step indicators to show active and completed states
    indicators.items.forEach((item, i) => {
      // Completed steps (steps before current)
      item.classList.toggle('step-completed', i < step - 1);
      // Current active step
      item.classList.toggle('step-active', i === step - 1);
      // Inactive steps (steps after current)
      item.classList.toggle('step-inactive', i > step - 1);
    });

    // Run validation for the current step
    validateStep(step);
    
    // Update button states based on validation
    updateNextButtonState(step);
    
    if (step === 2) updateTransformInfo(getVal(SELECTORS.transStrategy), $(SELECTORS.transformInfo));
  };

  // Only allow navigation to next step if validation passes
  btns.prev.addEventListener('click', () => current > 1 && show(--current));
  btns.next.addEventListener('click', () => {
    const stepKey = `step${current}`;
    if (validationState[stepKey] && current < total) {
      show(++current);
    } else {
      // Shake the button to indicate validation issues
      btns.next.classList.add('animate__animated', 'animate__headShake');
      setTimeout(() => {
        btns.next.classList.remove('animate__animated', 'animate__headShake');
      }, 500);
      
      // Force validation to show error messages
      validateStep(current, true);
    }
  });
  show(current);
}

// --- URL fields ---------------------------------------------------------
function initUrlFields() {
  const container = $(SELECTORS.urlsContainer);
  const addBtn = $(SELECTORS.addUrlBtn);

  const createField = () => {
    const div = document.createElement('div');
    div.className = 'lb-url field-container mb-2 animate__animated animate__fadeIn';
    div.innerHTML = `
      <div class="flex items-center">
        <input type="text" placeholder="https://" class="flex-1 border rounded-l-md py-2 px-3 focus:ring-orange-500">
        <button type="button" class="bg-gray-100 border rounded-r-md px-3 py-2 hover:bg-gray-200 remove-url">
          <i class="fa-solid fa-trash-alt text-gray-600"></i>
        </button>
      </div>`;
    
    // Add validation to the input field
    const input = div.querySelector('input');
    input.addEventListener('input', debounce(() => {
      validateUrlField(input);
      validateStep(3);
    }, TIME.VALIDATION_DEBOUNCE));
    
    attachRemove(div.querySelector('.remove-url'), div);
    return div;
  };

  const attachRemove = (btn, parent) => {
    btn.addEventListener('click', () => {
      parent.classList.replace('animate__fadeIn', 'animate__fadeOut');
      setTimeout(() => {
        parent.remove();
        validateStep(3); // Revalidate after removal
      }, TIME.URL_REMOVE_ANIM);
    });
  };

  addBtn.addEventListener('click', () => {
    const field = createField();
    container.appendChild(field);
    field.querySelector('input').focus();
  });
  
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
    
    // Validate all steps before submission
    validateStep1(true);
    validateStep2(true);
    validateStep3(true);
    validateStep4(true);
    
    if (!validationState.step4) {
      return showResponse('error', 'Please fix validation errors before submitting', resp);
    }
    
    const urls = $$('.lb-url input').map(i => i.value.trim()).filter(Boolean);
    if (!urls.length) {
      return showResponse('error', 'Please add at least one target URL', resp);
    }
    showResponse('loading', 'Processing request...', resp);

    const payload = {
      client_context: {
        channel: checkedVal('client_channel'),
        dial_timeout: 1e6 * Number(getVal('#dial-timeout')),
        dial_keep_alive: 1e6 * Number(getVal('#keepalive-timeout')),
        tls_handshake_timeout: 1e6 * Number(getVal('#tls-handshake-timeout')),
        response_header_timeout: 1e6 * Number(getVal('#response-header-timeout')),
        expect_continue_timeout: 1e6 * Number(getVal('#expect-continue-timeout')),
        request_timeout: 1e6 * Number(getVal('#request-timeout')),
        insecure_skip_verify: $('#insecure-skip-verify').checked
      },
      driver_context: {
        batch_size: Number(getVal('#batch-size')),
        should_store_responses: $('#store-responses').checked,
        responses_storage_path: $('#store-responses').checked ? getVal('#responses-path') : ""
      },
      parser_context: {
        strategy: checkedVal('parser_strategy'),
        file_path: getVal('#file-path'),
        file_content_b64: getVal('#file-content-b64')
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
        showResponse('success', 'Job started! Tracking progress...', resp);
        startJobPolling(data.id);
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
  
  // Get file info - either from file picker or text input
  let file = getVal('#file-path');
  const fileInput = document.getElementById('file-input');
  if (fileInput && fileInput.files && fileInput.files.length > 0) {
    // Use the actual filename if a file was selected
    file = fileInput.files[0].name;
  }
  
  const trans  = getVal('#trans-strategy');
  const method = getVal('#method-expr');
  const endpoint = getVal('#endpoint-expr');
  const headers = getVal('#headers-expr');
  const body = getVal('#body-expr');
  const clientChannel = checkedVal('client_channel');
  const dial = getVal('#dial-timeout');
  const keep = getVal('#keepalive-timeout');
  const tlsHandshake = getVal('#tls-handshake-timeout');
  const responseHeader = getVal('#response-header-timeout');
  const expectContinue = getVal('#expect-continue-timeout');
  const requestTimeout = getVal('#request-timeout');
  const insecureSkipVerify = $('#insecure-skip-verify').checked;
  const lbStrat = checkedVal('lb_strategy');
  const urlsArray = $$('.lb-url input').map(i => i.value.trim()).filter(Boolean);
  const batchSize = getVal('#batch-size');
  const storeResponses = $('#store-responses').checked;
  
  // Validate URLs
  const validUrls = urlsArray.filter(url => validateUrl(url));
  const invalidUrls = urlsArray.filter(url => !validateUrl(url) && url.trim() !== '');
  
  // Validate expressions
  const methodValid = validateExpression(method);
  const endpointValid = validateExpression(endpoint);
  const headersValid = validateExpression(headers);
  const bodyValid = validateExpression(body);
  
  // Generate URL list HTML
  let urlListHtml = '';
  if (urlsArray.length > 0) {
    urlListHtml = '<div class="urls-list">';
    urlsArray.forEach((url, index) => {
      const isValid = validateUrl(url);
      urlListHtml += `
        <div class="url-item">
          <i class="fas fa-link"></i>
          <span>${url}</span>
          ${isValid 
            ? '<span class="validation-indicator validation-success"><i class="fas fa-check"></i></span>' 
            : '<span class="validation-indicator validation-error"><i class="fas fa-times"></i></span>'}
        </div>
      `;
    });
    urlListHtml += '</div>';
  }

  summary.innerHTML = `
    <!-- Source Configuration Card -->
    <div class="config-card">
      <div class="config-card-header">
        <div class="icon-container source-icon-bg">
          <i class="fas fa-file-import"></i>
        </div>
        <h5>Source Configuration</h5>
      </div>
      <div class="config-card-body">
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-code"></i>
            Format
          </div>
          <div class="config-item-value">
            <span class="tag tag-blue">${escapeHtml(parser)}</span>
          </div>
        </div>
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-folder"></i>
            File Path
          </div>
          <div class="config-item-value">
            ${escapeHtml(file)}
          </div>
        </div>
      </div>
    </div>

    <!-- Transform Configuration Card -->
    <div class="config-card">
      <div class="config-card-header">
        <div class="icon-container transform-icon-bg">
          <i class="fas fa-sliders-h"></i>
        </div>
        <h5>Transform Configuration</h5>
      </div>
      <div class="config-card-body">
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-code-branch"></i>
            Strategy
          </div>
          <div class="config-item-value">
            <span class="tag tag-purple">${escapeHtml(trans)}</span>
          </div>
        </div>
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-terminal"></i>
            Method
          </div>
          <div class="config-item-value">
            ${escapeHtml(method)}
          </div>
        </div>
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-link"></i>
            Endpoint
          </div>
          <div class="config-item-value">
            ${escapeHtml(endpoint)}
          </div>
        </div>
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-tags"></i>
            Headers
          </div>
          <div class="config-item-value">
            Configured
          </div>
        </div>
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-file-alt"></i>
            Body
          </div>
          <div class="config-item-value">
            Configured
          </div>
        </div>
      </div>
    </div>

    <!-- Target Configuration Card -->
    <div class="config-card">
      <div class="config-card-header">
        <div class="icon-container target-icon-bg">
          <i class="fas fa-bullseye"></i>
        </div>
        <h5>Target Configuration</h5>
      </div>
      <div class="config-card-body">
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-plug"></i>
            Channel
          </div>
          <div class="config-item-value">
            <span class="tag tag-orange">${escapeHtml(clientChannel)}</span>
          </div>
        </div>
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-cogs"></i>
            Client
          </div>
          <div class="config-item-value">
            <button type="button" class="view-client-btn text-sm text-blue-600 hover:text-blue-800"
              onclick="toggleClientDetails()">View Details</button>
          </div>
        </div>

        <div id="timeout-details" class="config-item-details" style="display: none;">
          <div class="config-detail-item">
            <div class="detail-label">
              <i class="fas fa-clock"></i>
              Dial Timeout
            </div>
            <div class="detail-value">${escapeHtml(dial)} ms</div>
          </div>
          <div class="config-detail-item">
            <div class="detail-label">
              <i class="fas fa-heartbeat"></i>
              Keep Alive
            </div>
            <div class="detail-value">${escapeHtml(keep)} ms</div>
          </div>
          <div class="config-detail-item">
            <div class="detail-label">
              <i class="fas fa-shield-alt"></i>
              TLS Handshake
            </div>
            <div class="detail-value">${escapeHtml(tlsHandshake)} ms</div>
          </div>
          <div class="config-detail-item">
            <div class="detail-label">
              <i class="fas fa-file-code"></i>
              Response Header
            </div>
            <div class="detail-value">${escapeHtml(responseHeader)} ms</div>
          </div>
          <div class="config-detail-item">
            <div class="detail-label">
              <i class="fas fa-hourglass-half"></i>
              Expect-Continue
            </div>
            <div class="detail-value">${escapeHtml(expectContinue)} ms</div>
          </div>
          <div class="config-detail-item">
            <div class="detail-label">
              <i class="fas fa-stopwatch"></i>
              Request Timeout
            </div>
            <div class="detail-value">${escapeHtml(requestTimeout)} ms</div>
          </div>
          <div class="config-detail-item">
            <div class="detail-label">
              <i class="fas fa-lock${insecureSkipVerify ? '-open' : ''}"></i>
              TLS Verification
            </div>
            <div class="detail-value">${insecureSkipVerify ? 'Disabled' : 'Enabled'}</div>
          </div>
        </div>
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-balance-scale"></i>
            Load Balancer
          </div>
          <div class="config-item-value">
            <span class="tag tag-green">${escapeHtml(lbStrat)}</span>
          </div>
        </div>
        <div class="config-item">
          <div class="config-item-label">
            <i class="fas fa-server"></i>
            Target URLs
          </div>
          <div class="config-item-value">
            <span class="tag tag-blue">${urlsArray.length}</span> URLs configured
          </div>
        </div>
        ${urlsArray.length > 0 ? `
        <div class="config-item" style="display: block;">
          <div class="urls-list">
            ${urlsArray.map(url => `
              <div class="url-item">
                <i class="fas fa-link"></i>
                <span>${escapeHtml(url)}</span>
              </div>
            `).join('')}
          </div>
        </div>
        ` : ''}
      </div>
    </div>
  `;
}

// --- Transform Info ----------------------------------------------------
function updateTransformInfo(strategy, el) {
  if (!el) return;
  const text = TRANSFORM_INFO[strategy] || TRANSFORM_INFO.DEFAULT;
  el.innerHTML = `${ICON.INFO}<span>${text}</span>`;
}

function initTransformInfoListener() {
  const sel = $(SELECTORS.transStrategy);
  const info = $(SELECTORS.transformInfo);
  sel?.addEventListener('change', () => updateTransformInfo(sel.value, info));
}

// --- Timeout Details Toggle ---------------------------------------------
function toggleClientDetails() {
  const timeoutDetails = document.getElementById('timeout-details');
  if (timeoutDetails) {
    const isCurrentlyHidden = timeoutDetails.style.display === 'none';
    timeoutDetails.style.display = isCurrentlyHidden ? 'block' : 'none';
    
    // Update button text and ARIA attribute
    const btn = document.querySelector('.view-client-btn');
    if (btn) {
      btn.textContent = isCurrentlyHidden ? 'Hide Details' : 'View Details';
      btn.setAttribute('aria-expanded', isCurrentlyHidden ? 'true' : 'false');
    }
  }
}

// --- Configuration Issues -----------------------------------------------
function initConfigurationIssues() {
  const viewIssuesBtn = $('#view-issues-btn');
  const checkConfigBtn = $('#check-config-btn');
  const closeIssuesBtn = $('#close-issues-btn');
  const issuesSection = $('#issues-section');
  
  if (!viewIssuesBtn || !checkConfigBtn || !closeIssuesBtn || !issuesSection) return;
  
  // Show issues section when view-issues button is clicked
  viewIssuesBtn.addEventListener('click', () => {
    const issuesContainer = $('#configuration-issues');
    if (!issuesContainer.innerHTML) {
      displayConfigurationIssues();
    }
    issuesSection.classList.remove('hidden');
  });
  
  // Refresh issues when check-config button is clicked
  checkConfigBtn.addEventListener('click', () => {
    displayConfigurationIssues();
  });
  
  // Hide issues section when close-issues button is clicked
  closeIssuesBtn.addEventListener('click', () => {
    issuesSection.classList.add('hidden');
  });
  
  // Check for issues when reaching the review step
  const steps = $$(SELECTORS.steps);
  if (steps.length >= 4) {
    const reviewStep = steps[3]; // Fourth step (0-indexed)
    const observer = new MutationObserver(function(mutations) {
      mutations.forEach(function(mutation) {
        if (mutation.attributeName === 'class' && 
            !reviewStep.classList.contains('hidden') && 
            reviewStep.classList.contains('visible')) {
          displayConfigurationIssues();
        }
      });
    });
    
    observer.observe(reviewStep, { attributes: true });
  }
}

// --- Configuration Issues ------------------------------------------------
function displayConfigurationIssues() {
  // Collect all validation issues
  const issues = [];
  
  // Check source configuration
  const filePath = getVal('#file-path');
  if (!validateFilePath(filePath)) {
    issues.push({ category: 'Source', message: 'Invalid file path format', severity: 'error' });
  }
  
  // Check transform configuration
  const methodExpr = getVal('#method-expr');
  const endpointExpr = getVal('#endpoint-expr');
  const headersExpr = getVal('#headers-expr');
  const bodyExpr = getVal('#body-expr');
  
  if (!validateExpression(methodExpr)) {
    issues.push({ category: 'Transform', message: 'Invalid method expression', severity: 'error' });
  }
  if (!validateExpression(endpointExpr)) {
    issues.push({ category: 'Transform', message: 'Invalid endpoint expression', severity: 'error' });
  }
  if (!validateExpression(headersExpr)) {
    issues.push({ category: 'Transform', message: 'Headers expression has unbalanced quotes or brackets', severity: 'error' });
  }
  if (!validateExpression(bodyExpr)) {
    issues.push({ category: 'Transform', message: 'Body expression has unbalanced quotes or brackets', severity: 'error' });
  }
  
  // Check target configuration
  const dialTimeout = getVal('#dial-timeout');
  const keepaliveTimeout = getVal('#keepalive-timeout');
  const urlInputs = $$('.lb-url input');
  const urlsArray = urlInputs.map(i => i.value.trim()).filter(Boolean);
  
  if (!validateNumberRange(dialTimeout, VALIDATION.MIN_TIMEOUT, VALIDATION.MAX_TIMEOUT)) {
    issues.push({ 
      category: 'Target', 
      message: `Dial timeout must be between ${VALIDATION.MIN_TIMEOUT} and ${VALIDATION.MAX_TIMEOUT} ms`,
      severity: 'error' 
    });
  }
  
  if (!validateNumberRange(keepaliveTimeout, VALIDATION.MIN_TIMEOUT, VALIDATION.MAX_TIMEOUT)) {
    issues.push({ 
      category: 'Target', 
      message: `Keep alive timeout must be between ${VALIDATION.MIN_TIMEOUT} and ${VALIDATION.MAX_TIMEOUT} ms`,
      severity: 'error' 
    });
  }
  
  if (urlsArray.length === 0) {
    issues.push({ category: 'Target', message: 'No target URLs configured', severity: 'error' });
  } else {
    urlsArray.forEach((url, index) => {
      if (!validateUrl(url)) {
        issues.push({ 
          category: 'Target', 
          message: `Invalid URL format at position ${index + 1}: ${url}`,
          severity: 'error' 
        });
      }
    });
  }
  
  // We don't check driver configuration in the summary view 
  // since it's already shown separately on the review screen
  
  // Display issues
  const issuesContainer = $('#configuration-issues');
  if (!issuesContainer) return;
  
  // Clear previous issues
  issuesContainer.innerHTML = '';
  
  if (issues.length === 0) {
    issuesContainer.innerHTML = `
      <div class="flex items-center p-4 bg-green-50 text-green-700 rounded-lg mb-4 animate__animated animate__fadeIn">
        <i class="fas fa-check-circle text-xl mr-3"></i>
        <p class="font-medium">No configuration issues found. You're ready to run the bombardment!</p>
      </div>
    `;
    $('#issue-count').textContent = '0';
    $('#issue-badge').classList.add('hidden');
    return;
  }
  
  // Update the issue count
  $('#issue-count').textContent = issues.length;
  $('#issue-badge').classList.remove('hidden');
  
  // Group issues by category
  const issuesByCategory = issues.reduce((acc, issue) => {
    if (!acc[issue.category]) {
      acc[issue.category] = [];
    }
    acc[issue.category].push(issue);
    return acc;
  }, {});
  
  // Create issues list with expandable sections
  const errorCount = issues.filter(i => i.severity === 'error').length;
  const warningCount = issues.filter(i => i.severity === 'warning').length;
  
  // Add summary header
  issuesContainer.innerHTML = `
    <div class="mb-4 p-3 bg-gray-50 rounded-lg border border-gray-200 flex items-center justify-between">
      <div class="flex items-center">
        <i class="fas fa-exclamation-triangle text-orange-500 mr-2"></i>
        <h3 class="font-medium">Configuration Issues Found</h3>
      </div>
      <div class="flex items-center gap-2">
        ${errorCount > 0 ? `<span class="px-2 py-1 bg-red-100 text-red-800 rounded-full text-xs font-semibold">${errorCount} ${errorCount === 1 ? 'Error' : 'Errors'}</span>` : ''}
        ${warningCount > 0 ? `<span class="px-2 py-1 bg-yellow-100 text-yellow-800 rounded-full text-xs font-semibold">${warningCount} ${warningCount === 1 ? 'Warning' : 'Warnings'}</span>` : ''}
      </div>
    </div>
  `;
  
  // Create accordion for each category
  Object.keys(issuesByCategory).forEach((category, index) => {
    const categoryIssues = issuesByCategory[category];
    const hasErrors = categoryIssues.some(i => i.severity === 'error');
    
    const categorySection = document.createElement('div');
    categorySection.className = 'mb-3 border border-gray-200 rounded-lg overflow-hidden animate__animated animate__fadeIn';
    categorySection.style.animationDelay = `${index * 0.1}s`;
    
    // Create header
    const header = document.createElement('div');
    header.className = `flex items-center justify-between p-3 cursor-pointer ${hasErrors ? 'bg-red-50' : 'bg-gray-50'}`;
    header.innerHTML = `
      <div class="flex items-center">
        <i class="fas fa-${
          category === 'Source' ? 'file-import' : 
          category === 'Transform' ? 'sliders-h' : 
          category === 'Target' ? 'bullseye' : 
          'cog'
        } mr-2 ${hasErrors ? 'text-red-500' : 'text-gray-700'}"></i>
        <h4 class="font-medium">${category} Configuration</h4>
      </div>
      <div class="flex items-center">
        <span class="mr-2 text-sm text-gray-500">${categoryIssues.length} ${categoryIssues.length === 1 ? 'issue' : 'issues'}</span>
        <i class="fas fa-chevron-down category-toggle-icon transition-transform"></i>
      </div>
    `;
    
    // Create issues list
    const issuesList = document.createElement('div');
    issuesList.className = 'category-issues hidden';
    
    categoryIssues.forEach(issue => {
      const issueElement = document.createElement('div');
      issueElement.className = `p-3 border-t border-gray-200 ${
        issue.severity === 'error' ? 'bg-red-50' : 
        issue.severity === 'warning' ? 'bg-yellow-50' : 
        'bg-blue-50'
      }`;
      
      issueElement.innerHTML = `
        <div class="flex items-center justify-between">
          <div class="flex items-center">
            <i class="mr-2 fas fa-${
              issue.severity === 'error' ? 'times-circle text-red-500' : 
              issue.severity === 'warning' ? 'exclamation-circle text-yellow-500' : 
              'info-circle text-blue-500'
            }"></i>
            <span>${issue.message}</span>
          </div>
          <span class="px-2 py-1 rounded-full text-xs font-semibold uppercase ${
            issue.severity === 'error' ? 'bg-red-100 text-red-800' : 
            issue.severity === 'warning' ? 'bg-yellow-100 text-yellow-800' : 
            'bg-blue-100 text-blue-800'
          }">${issue.severity}</span>
        </div>
      `;
      
      issuesList.appendChild(issueElement);
    });
    
    categorySection.appendChild(header);
    categorySection.appendChild(issuesList);
    
    // Toggle functionality
    header.addEventListener('click', () => {
      const icon = header.querySelector('.category-toggle-icon');
      const content = header.nextElementSibling;
      
      icon.classList.toggle('rotate-180');
      content.classList.toggle('hidden');
      
      // Auto-expand the first section
      if (index === 0 && content.classList.contains('hidden')) {
        content.classList.remove('hidden');
        icon.classList.add('rotate-180');
      }
    });
    
    issuesContainer.appendChild(categorySection);
    
    // Auto-expand the first section with errors
    if ((index === 0 || hasErrors) && categoryIssues.length > 0) {
      header.click();
    }
  });
  
  // Add helpful tip at the bottom
  const tipElement = document.createElement('div');
  tipElement.className = 'mt-4 text-sm text-gray-600 italic flex items-center animate__animated animate__fadeIn animate__delay-1s';
  tipElement.innerHTML = `
    <i class="fas fa-lightbulb mr-1 text-yellow-500"></i>
    Tip: Fix the issues in each category and then click "Check Configuration" to verify your changes.
  `;
  issuesContainer.appendChild(tipElement);
  
  // Reveal issues section if hidden
  $('#issues-section').classList.remove('hidden');
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

// --- File Picker Functionality ------------------------------------------
function initFileInput() {
  const fileInput = $(SELECTORS.fileInput);
  const filePathDisplay = $(SELECTORS.filePathDisplay);
  const filePathHidden = $(SELECTORS.filePathHidden);
  const fileDetails = $(SELECTORS.fileDetails);
  const fileSizeEl = $(SELECTORS.fileSize);
  const fileModifiedEl = $(SELECTORS.fileModified);
  const filePathNote = $(SELECTORS.filePathNote);
  const fileContentB64 = $(SELECTORS.fileContentB64);
  
  if (!fileInput) return;
  
  // Handle clicking on the file path display to trigger file input
  filePathDisplay.addEventListener('click', () => {
    fileInput.click();
  });
  
  // Handle keyboard interactions for accessibility
  filePathDisplay.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      fileInput.click();
    }
  });
  
  fileInput.addEventListener('change', (e) => {
    const file = e.target.files[0];
    
    if (file) {
      // Display file name
      filePathDisplay.value = file.name;
      
      // Store the file path in a hidden field
      const filePath = `./data/${file.name}`;
      filePathHidden.value = filePath;
      
      // Show file details
      const fileSizeFormatted = formatFileSize(file.size);
      const fileModified = new Date(file.lastModified).toLocaleDateString();
      
      fileSizeEl.textContent = fileSizeFormatted;
      fileModifiedEl.textContent = `Modified: ${fileModified}`;
      filePathNote.textContent = `File will be uploaded as: ${filePath}`;
      fileDetails.classList.remove('hidden');
      
      // Read file content as base64 to send with the request
      const reader = new FileReader();
      reader.onload = function(e) {
        const base64Content = e.target.result.split(',')[1]; // Remove the data:*/* prefix
        fileContentB64.value = base64Content;
        console.log("File content encoded as base64", { size: base64Content.length });
      };
      reader.readAsDataURL(file);
      
      // Trigger validation
      validateStep1();
    } else {
      // Clear the display if no file selected
      filePathDisplay.value = '';
      filePathHidden.value = '';
      fileContentB64.value = '';
      fileDetails.classList.add('hidden');
    }
  });
}

// Update file picker accept attribute based on parser strategy
function updateFileAccept() {
  const fileInput = $(SELECTORS.fileInput);
  const parserStrategy = checkedVal('parser_strategy');
  
  if (!fileInput) return;
  
  // Set accepted file types based on parser strategy
  switch (parserStrategy) {
    case 'CSV':
      fileInput.setAttribute('accept', '.csv');
      break;
    case 'JSON':
      fileInput.setAttribute('accept', '.json');
      break;
    case 'XML':
      fileInput.setAttribute('accept', '.xml');
      break;
    case 'YAML':
      fileInput.setAttribute('accept', '.yaml,.yml');
      break;
    default:
      fileInput.setAttribute('accept', '.csv,.json,.xml,.yaml,.yml');
  }
}

// Helper function to format file size
function formatFileSize(bytes) {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Initialize parser strategy change listeners for file picker
function initParserStrategyListeners() {
  const parserRadios = document.querySelectorAll('input[name="parser_strategy"]');
  if (!parserRadios.length) return;
  
  parserRadios.forEach(radio => {
    radio.addEventListener('change', () => {
      updateFileAccept();
    });
  });
  
  // Set initial file accept attribute
  updateFileAccept();
}

// --- Store Responses Toggle ---------------------------------------------
function initStoreResponsesListener() {
  const storeResponsesCheckbox = $('#store-responses');
  const responsesPathContainer = $('#responses-path-container');
  
  if (!storeResponsesCheckbox || !responsesPathContainer) return;

  // Initialize visibility based on the current checked state
  if (storeResponsesCheckbox.checked) {
    responsesPathContainer.classList.remove('hidden');
  } else {
    responsesPathContainer.classList.add('hidden');
  }
  
  storeResponsesCheckbox.addEventListener('change', function() {
    if (this.checked) {
      responsesPathContainer.classList.remove('hidden');
    } else {
      responsesPathContainer.classList.add('hidden');
    }
  });
}

// --- Job Progress Polling ------------------------------------------------

let activePollingInterval = null;

function startJobPolling(jobId) {
  const progressEl = $('#job-progress');
  if (progressEl) progressEl.classList.remove('hidden');

  updateProgressDisplay({
    id: jobId,
    status: 'PENDING',
    progress_percent: 0,
    processed_rows: 0,
    failed_rows: 0,
    total_rows: 0,
    error_message: ''
  });

  if (activePollingInterval) clearInterval(activePollingInterval);

  activePollingInterval = setInterval(async () => {
    try {
      const res = await fetch('/v1/bombardment/' + encodeURIComponent(jobId));
      if (!res.ok) return;
      const job = await res.json();
      updateProgressDisplay(job);

      if (job.status === 'COMPLETED' || job.status === 'FAILED') {
        clearInterval(activePollingInterval);
        activePollingInterval = null;
      }
    } catch (err) {
      console.error('Polling error:', err);
    }
  }, 1500);
}

function updateProgressDisplay(job) {
  const setTextSafe = (sel, text) => {
    const el = $(sel);
    if (el) el.textContent = String(text);
  };

  setTextSafe('#job-progress-id', 'Job: ' + (job.id || '').substring(0, 8) + '...');

  const statusEl = $('#job-progress-status');
  if (statusEl) {
    statusEl.textContent = job.status;
    statusEl.className = 'tag';
    switch (job.status) {
      case 'COMPLETED':
        statusEl.classList.add('tag-green');
        break;
      case 'FAILED':
        statusEl.classList.add('tag-red');
        break;
      case 'RUNNING':
        statusEl.classList.add('tag-orange');
        break;
      default:
        statusEl.classList.add('tag-blue');
    }
  }

  const pct = Math.min(100, Math.max(0, job.progress_percent || 0));
  setTextSafe('#job-progress-percent', pct.toFixed(1) + '%');
  const bar = $('#job-progress-bar');
  if (bar) {
    bar.style.width = pct + '%';
    bar.classList.remove('completed', 'failed');
    if (job.status === 'COMPLETED') bar.classList.add('completed');
    if (job.status === 'FAILED') bar.classList.add('failed');
  }

  setTextSafe('#job-progress-processed', job.processed_rows || 0);
  setTextSafe('#job-progress-failed', job.failed_rows || 0);
  setTextSafe('#job-progress-total', job.total_rows || 0);

  const errEl = $('#job-progress-error');
  if (errEl) {
    if (job.error_message) {
      errEl.textContent = job.error_message;
      errEl.classList.remove('hidden');
    } else {
      errEl.classList.add('hidden');
    }
  }

  const newBtn = $('#job-new-btn');
  if (newBtn) {
    if (job.status === 'COMPLETED' || job.status === 'FAILED') {
      newBtn.classList.remove('hidden');
    } else {
      newBtn.classList.add('hidden');
    }
  }
}

function initJobProgressActions() {
  const newBtn = $('#job-new-btn');
  if (newBtn) {
    newBtn.addEventListener('click', () => {
      const progressEl = $('#job-progress');
      if (progressEl) progressEl.classList.add('hidden');
      $(SELECTORS.respEl).innerHTML = '';
    });
  }
}
