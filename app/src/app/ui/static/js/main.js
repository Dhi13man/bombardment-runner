// UI logic for Bombardment form

document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('bombard-form');
  const respEl = document.getElementById('response-message');

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    respEl.textContent = '';

    // Client context
    const clientChannel = document.querySelector('input[name="client_channel"]:checked').value;
    const dialTimeout = parseInt(document.getElementById('dial-timeout').value, 10);
    const keepAlive = parseInt(document.getElementById('keepalive-timeout').value, 10);

    // Driver context
    const batchSize = parseInt(document.getElementById('batch-size').value, 10);
    const storeResponses = document.getElementById('store-responses').checked;

    // Parser context
    const parserStrategy = document.querySelector('input[name="parser_strategy"]:checked').value;
    const filePath = document.getElementById('file-path').value;

    // Load balancer
    const lbStrategy = document.querySelector('input[name="lb_strategy"]:checked').value;
    const urlInputs = document.querySelectorAll('.lb-url input');
    const urls = Array.from(urlInputs).map(i => i.value.trim()).filter(u => u);

    // Transformer
    const transStrategy = document.getElementById('trans-strategy').value;
    const methodExpr = document.getElementById('method-expr').value;
    const endpointExpr = document.getElementById('endpoint-expr').value;
    const headersExpr = document.getElementById('headers-expr').value;
    const bodyExpr = document.getElementById('body-expr').value;

    const payload = {
      client_context: {
        channel: clientChannel,
        dial_timeout: dialTimeout,
        keep_alive_timeout: keepAlive
      },
      driver_context: {
        batch_size: batchSize,
        should_store_responses: storeResponses
      },
      parser_context: {
        strategy: parserStrategy,
        file_path: filePath
      },
      load_balancer_context: {
        strategy: lbStrategy,
        urls: urls
      },
      transformer_context: {
        strategy: transStrategy,
        method_expression: methodExpr,
        endpoint_expression: endpointExpr,
        headers_expression: headersExpr,
        body_expression: bodyExpr
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
        respEl.textContent = 'Bombardment started!';
        respEl.className = 'text-green-600';
      } else {
        respEl.textContent = data.error || 'Error starting bombardment';
        respEl.className = 'text-red-600';
      }
    } catch (err) {
      respEl.textContent = err.message;
      respEl.className = 'text-red-600';
    }
  });
});
