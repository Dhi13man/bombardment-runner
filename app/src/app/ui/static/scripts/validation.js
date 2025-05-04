// Validation helpers for the bombardment UI

// Access selectors from main.js
/* global SELECTORS */

// Validate a text field and show/clear error
function validateTextField(input, validationFn, errorMessage) {
  const value = input.value.trim();
  const isValid = validationFn(value);
  
  // Remove any existing error message
  const existingError = input.parentElement.querySelector('.validation-error');
  if (existingError) {
    existingError.remove();
  }
  
  // Clear existing styling
  input.classList.remove('border-red-500', 'border-green-500', 'focus:border-red-500', 'focus:border-green-500');
  input.classList.remove('focus:ring-red-500', 'focus:ring-green-500');
  
  if (!isValid) {
    // Add error styling and message
    input.classList.add('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
    
    // Add error message
    const errorEl = document.createElement('div');
    errorEl.className = 'validation-error text-xs text-red-500 mt-1 flex items-center';
    errorEl.innerHTML = `<i class="fas fa-exclamation-circle mr-1"></i>${errorMessage}`;
    input.parentElement.appendChild(errorEl);
    return false;
  } else if (value.length > 0) {
    // Add success styling for non-empty inputs
    input.classList.add('border-green-500', 'focus:border-green-500', 'focus:ring-green-500');
    return true;
  }
  
  return value.length === 0 ? false : true; // Empty inputs are invalid unless explicitly allowed
}

// Validate URL
function validateUrl(url) {
  if (!url || url.length === 0) return false;
  return VALIDATION.URL_REGEX.test(url);
}

// Validate file path
function validateFilePath(path) {
  if (!path || path.length === 0) return false;
  return VALIDATION.FILE_PATH_REGEX.test(path);
}

// Validate expression (basic check - not empty and has valid quotes/brackets)
function validateExpression(expr) {
  if (!expr || expr.length === 0) return false;
  
  // Check for balanced quotes
  const singleQuotes = (expr.match(/'/g) || []).length;
  const doubleQuotes = (expr.match(/"/g) || []).length;
  if (singleQuotes % 2 !== 0 || doubleQuotes % 2 !== 0) return false;
  
  // Check for balanced brackets
  const openBraces = (expr.match(/{/g) || []).length;
  const closeBraces = (expr.match(/}/g) || []).length;
  const openBrackets = (expr.match(/\[/g) || []).length;
  const closeBrackets = (expr.match(/\]/g) || []).length;
  const openParens = (expr.match(/\(/g) || []).length;
  const closeParens = (expr.match(/\)/g) || []).length;
  
  return openBraces === closeBraces && 
         openBrackets === closeBrackets && 
         openParens === closeParens;
}

// Validate number is within range
function validateNumberRange(value, min, max) {
  const num = Number(value);
  return !isNaN(num) && num >= min && num <= max;
}

// Validate a URL field
function validateUrlField(input) {
  return validateTextField(
    input,
    validateUrl,
    'Please enter a valid URL (e.g., https://example.com)'
  );
}

// Add feedback container below an element
function addFeedbackContainer(elementId, containerId) {
  const element = document.getElementById(elementId);
  if (!element) return null;
  
  // Check if container already exists
  let container = document.getElementById(containerId);
  if (container) return container;
  
  // Create new container
  container = document.createElement('div');
  container.id = containerId;
  container.className = 'mt-2 text-sm';
  
  // Insert after the element
  element.parentNode.insertBefore(container, element.nextSibling);
  return container;
}

// Show validation message in a container
function showValidationMessage(containerId, isValid, message) {
  const container = document.getElementById(containerId);
  if (!container) return;
  
  const iconType = isValid ? 'SUCCESS' : 'ERROR';
  const textColorClass = isValid ? 'text-green-600' : 'text-red-500';
  
  container.className = `mt-2 text-sm ${textColorClass}`;
  container.innerHTML = `${ICON[iconType]}<span>${message}</span>`;
}

// Clear validation message
function clearValidationMessage(containerId) {
  const container = document.getElementById(containerId);
  if (container) {
    container.innerHTML = '';
  }
}

// Update button state based on validation
function updateNextButtonState(step) {
  const stepKey = `step${step}`;
  const nextBtn = document.querySelector(step === 4 ? SELECTORS.submitBtn : SELECTORS.nextBtn);
  
  if (nextBtn) {
    if (validationState[stepKey]) {
      nextBtn.classList.remove('opacity-50', 'cursor-not-allowed');
      nextBtn.classList.add('hover:opacity-90');
      nextBtn.disabled = false;
    } else {
      nextBtn.classList.add('opacity-50', 'cursor-not-allowed');
      nextBtn.classList.remove('hover:opacity-90');
      nextBtn.disabled = true;
    }
  }
}

// Validate Step 1 - Source Configuration
function validateStep1(showErrors = false) {
  const filePathInput = document.getElementById('file-path');
  const filePathValid = validateTextField(
    filePathInput, 
    validateFilePath,
    'Please enter a valid file path (e.g., ./data.csv)'
  );
  
  // Add additional source validations here if needed
  
  validationState.step1 = filePathValid;
  updateNextButtonState(1);
  return filePathValid;
}

// Validate Step 2 - Transform Configuration
function validateStep2(showErrors = false) {
  const methodInput = document.getElementById('method-expr');
  const endpointInput = document.getElementById('endpoint-expr');
  const headersInput = document.getElementById('headers-expr');
  const bodyInput = document.getElementById('body-expr');
  
  const methodValid = validateTextField(
    methodInput,
    validateExpression,
    'Method expression is invalid'
  );
  
  const endpointValid = validateTextField(
    endpointInput,
    validateExpression,
    'Endpoint expression is invalid'
  );
  
  const headersValid = validateTextField(
    headersInput,
    validateExpression,
    'Headers expression has unbalanced quotes or brackets'
  );
  
  const bodyValid = validateTextField(
    bodyInput,
    validateExpression,
    'Body expression has unbalanced quotes or brackets'
  );
  
  validationState.step2 = methodValid && endpointValid && headersValid && bodyValid;
  updateNextButtonState(2);
  return validationState.step2;
}

// Validate Step 3 - Target Configuration
function validateStep3(showErrors = false) {
  // Validate all URL fields
  const urlInputs = document.querySelectorAll('.lb-url input');
  let urlsValid = urlInputs.length > 0; // We need at least one URL field
  
  urlInputs.forEach(input => {
    const isValid = validateUrlField(input);
    urlsValid = urlsValid && (isValid || input.value.trim() === '');
  });
  
  // Check if at least one valid URL exists
  const hasValidUrl = Array.from(urlInputs).some(input => 
    input.value.trim() !== '' && validateUrl(input.value.trim())
  );
  
  // Validate timeouts
  const dialTimeoutInput = document.getElementById('dial-timeout');
  const keepaliveTimeoutInput = document.getElementById('keepalive-timeout');
  
  const dialTimeoutValid = validateTextField(
    dialTimeoutInput,
    val => validateNumberRange(val, VALIDATION.MIN_TIMEOUT, VALIDATION.MAX_TIMEOUT),
    `Dial timeout must be between ${VALIDATION.MIN_TIMEOUT} and ${VALIDATION.MAX_TIMEOUT} ms`
  );
  
  const keepaliveTimeoutValid = validateTextField(
    keepaliveTimeoutInput,
    val => validateNumberRange(val, VALIDATION.MIN_TIMEOUT, VALIDATION.MAX_TIMEOUT),
    `Keep alive timeout must be between ${VALIDATION.MIN_TIMEOUT} and ${VALIDATION.MAX_TIMEOUT} ms`
  );
  
  // Show URL list validation message if necessary
  const urlsContainer = document.getElementById('urls-container');
  const urlsMessageId = 'urls-validation-message';
  const urlsMessageContainer = addFeedbackContainer('urls-container', urlsMessageId);
  
  if (!hasValidUrl && showErrors) {
    showValidationMessage(urlsMessageId, false, 'At least one valid URL is required');
  } else if (hasValidUrl) {
    showValidationMessage(urlsMessageId, true, 'URLs are valid');
  } else {
    clearValidationMessage(urlsMessageId);
  }
  
  validationState.step3 = urlsValid && hasValidUrl && dialTimeoutValid && keepaliveTimeoutValid;
  updateNextButtonState(3);
  return validationState.step3;
}

// Validate Step 4 - Review Configuration
function validateStep4(showErrors = false) {
  const batchSizeInput = document.getElementById('batch-size');
  
  const batchSizeValid = validateTextField(
    batchSizeInput,
    val => validateNumberRange(val, VALIDATION.MIN_BATCH_SIZE, VALIDATION.MAX_BATCH_SIZE),
    `Batch size must be between ${VALIDATION.MIN_BATCH_SIZE} and ${VALIDATION.MAX_BATCH_SIZE}`
  );
  
  // Check other steps are also valid
  const allPreviousStepsValid = validationState.step1 && validationState.step2 && validationState.step3;
  
  validationState.step4 = batchSizeValid && allPreviousStepsValid;
  updateNextButtonState(4);
  return validationState.step4;
}

// Validate a specific step
function validateStep(stepNumber, showErrors = false) {
  switch (stepNumber) {
    case 1: return validateStep1(showErrors);
    case 2: return validateStep2(showErrors);
    case 3: return validateStep3(showErrors);
    case 4: return validateStep4(showErrors);
    default: return false;
  }
}

// Initialize all validations
function initValidation() {
  // Add event listeners to form fields for step 1
  document.getElementById('file-path')?.addEventListener('input', debounce(() => {
    validateStep1();
  }, TIME.VALIDATION_DEBOUNCE));
  
  // Add event listeners to form fields for step 2
  document.getElementById('method-expr')?.addEventListener('input', debounce(() => {
    validateStep2();
  }, TIME.VALIDATION_DEBOUNCE));
  
  document.getElementById('endpoint-expr')?.addEventListener('input', debounce(() => {
    validateStep2();
  }, TIME.VALIDATION_DEBOUNCE));
  
  document.getElementById('headers-expr')?.addEventListener('input', debounce(() => {
    validateStep2();
  }, TIME.VALIDATION_DEBOUNCE));
  
  document.getElementById('body-expr')?.addEventListener('input', debounce(() => {
    validateStep2();
  }, TIME.VALIDATION_DEBOUNCE));
  
  // Add event listeners to form fields for step 3
  document.getElementById('dial-timeout')?.addEventListener('input', debounce(() => {
    validateStep3();
  }, TIME.VALIDATION_DEBOUNCE));
  
  document.getElementById('keepalive-timeout')?.addEventListener('input', debounce(() => {
    validateStep3();
  }, TIME.VALIDATION_DEBOUNCE));
  
  // Add event listeners to form fields for step 4
  document.getElementById('batch-size')?.addEventListener('input', debounce(() => {
    validateStep4();
  }, TIME.VALIDATION_DEBOUNCE));
  
  // Initial validation of step 1
  validateStep1();
}
