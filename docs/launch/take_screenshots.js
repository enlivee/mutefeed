const { spawn } = require('child_process');
const fs = require('fs');

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function run() {
  const chrome = spawn('/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', [
    '--headless',
    '--disable-gpu',
    '--remote-debugging-port=9222',
    '--window-size=1270,760',
    'about:blank'
  ]);

  await sleep(1500);

  const res = await fetch('http://127.0.0.1:9222/json/list');
  const tabs = await res.json();
  const wsUrl = tabs[0].webSocketDebuggerUrl;

  const ws = new WebSocket(wsUrl);
  let id = 1;
  const pending = new Map();

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.id && pending.has(data.id)) {
      pending.get(data.id)(data);
      pending.delete(data.id);
    }
  };

  await new Promise(resolve => ws.onopen = resolve);

  function send(method, params = {}) {
    return new Promise(resolve => {
      const msgId = id++;
      pending.set(msgId, resolve);
      ws.send(JSON.stringify({ id: msgId, method, params }));
    });
  }

  await send('Page.enable');
  await send('Runtime.enable');
  await send('DOM.enable');

  // Navigate to mutefeed
  console.log('Navigating to mutefeed...');
  await send('Page.navigate', { url: 'https://mutefeed.onrender.com' });
  await sleep(4000);

  async function takeScreenshot(filepath) {
    const shot = await send('Page.captureScreenshot', { format: 'png' });
    fs.writeFileSync(filepath, Buffer.from(shot.result.data, 'base64'));
    console.log(`Saved: ${filepath}`);
  }

  // Slide 1: Main Clean Feed (HN filtered)
  await takeScreenshot('docs/launch/gallery-1-clean-feed.png');

  // Slide 2: Switch to Clean Tech Digest
  console.log('Switching to Clean Tech Digest...');
  await send('Runtime.evaluate', {
    expression: `
      const buttons = Array.from(document.querySelectorAll('button'));
      const digestBtn = buttons.find(b => b.textContent.includes('Clean Tech Digest') || b.textContent.includes('Smart Multi-Feed'));
      if (digestBtn) digestBtn.click();
    `
  });
  await sleep(3500);
  await takeScreenshot('docs/launch/gallery-2-tech-digest.png');

  // Slide 3: Open OPML Modal
  console.log('Opening Batch OPML modal...');
  await send('Runtime.evaluate', {
    expression: `
      const buttons = Array.from(document.querySelectorAll('button'));
      const opmlBtn = buttons.find(b => b.textContent.includes('Batch OPML') || b.textContent.includes('OPML'));
      if (opmlBtn) opmlBtn.click();
    `
  });
  await sleep(1000);
  await takeScreenshot('docs/launch/gallery-3-opml-cleaner.png');

  // Close OPML modal and open Reader integration modal
  await send('Runtime.evaluate', {
    expression: `
      const escEvent = new KeyboardEvent('keydown', { key: 'Escape', code: 'Escape', keyCode: 27, which: 27 });
      window.dispatchEvent(escEvent);
    `
  });
  await sleep(500);

  console.log('Opening Reader modal...');
  await send('Runtime.evaluate', {
    expression: `
      const buttons = Array.from(document.querySelectorAll('button'));
      const appBtn = buttons.find(b => b.textContent.includes('Open in App') || b.textContent.includes('Subscribe'));
      if (appBtn) appBtn.click();
    `
  });
  await sleep(1000);
  await takeScreenshot('docs/launch/gallery-4-reader-integration.png');

  ws.close();
  chrome.kill();
  console.log('Done!');
}

run().catch(err => {
  console.error(err);
  process.exit(1);
});
