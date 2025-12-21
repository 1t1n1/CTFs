const crypto = require('crypto');
const fs = require('fs');
const puppeteer = require('puppeteer');
const express = require('express');

const app = express();
const port = 3000;

app.use(express.urlencoded({ extended: false }));

function getPasswordHash(filePath, fallback) {
    try {
        const content = fs.readFileSync(filePath, 'utf8').trim();
        return crypto.createHash('sha256').update(content).digest('hex');
    } catch (error) {
        return fallback;
    }
}

const APP_URL = 'http://proxy';

// Browser instance management
let browserInstance = null;
const MAX_CONCURRENT = 3;
let activeVisits = 0;
const queue = [];

async function getBrowser() {
  if (!browserInstance || !browserInstance.isConnected()) {
    console.log('Launching new browser instance...');
    browserInstance = await puppeteer.launch({
      headless: "new",
      executablePath: "/usr/bin/chromium",
      args: [
        "--no-sandbox",
        "--disable-dev-shm-usage",
        "--disable-gpu",
        '--js-flags="--noexpose_wasm"',
      ],
    });
  }
  return browserInstance;
}

const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));

const processQueue = async () => {
  if (activeVisits >= MAX_CONCURRENT || queue.length === 0) {
    return;
  }

  activeVisits++;
  const task = queue.shift();

  try {
    await visit(task.id, task.type, task.productId);
  } catch (err) {
    console.error('Visit error:', err);
  } finally {
    activeVisits--;
    processQueue();
  }
};

const visit = async (id, type, productId) => {
  console.log(`start: ${id} (type: ${type}, productId: ${productId})`);

  const browser = await getBrowser();
  const context = await browser.createBrowserContext();

  try {
    const page = await context.newPage();

    await page.goto(`${APP_URL}/account/login`, { timeout: 3_000 });
    await page.waitForSelector('form', { timeout: 3_000 });

    await page.type('input[name="email"]', 'admin@example.com');
    await page.type('input[name="password"]', getPasswordHash('/flag1', 'admin'));
    await Promise.all([
      page.click('button[type="submit"]'),
      page.waitForNavigation({ timeout: 3_000 }),
    ]);

    await page.goto(`${APP_URL}/inquiry/${id}`, { timeout: 3_000 });
    await page.waitForSelector('form', { timeout: 3_000 });

    // Use the inquiry type passed from the server
    const isRestockRequest = type === 'restock';

    let responseMessage = 'Hello, I am a bot.';
    if (isRestockRequest && productId) {
      // Restock the product
      await page.goto(`${APP_URL}/admin/products`, { timeout: 3_000 });
      await page.waitForSelector('form', { timeout: 3_000 });

      // Fill in the restock form
      await page.evaluate((pid) => {
        const productIdInput = document.querySelector(`input[name="productId"][value="${pid}"]`);
        if (productIdInput) {
          const form = productIdInput.closest('form');
          const stockInput = form.querySelector('input[name="stockAmount"]');
          if (stockInput) {
            stockInput.value = '10';
          }
        }
      }, productId);

      // Submit the restock form
      await page.evaluate((pid) => {
        const productIdInput = document.querySelector(`input[name="productId"][value="${pid}"]`);
        if (productIdInput) {
          productIdInput.closest('form').submit();
        }
      }, productId);

      await page.waitForNavigation({ timeout: 3_000 });

      // Go back to inquiry and respond
      await page.goto(`${APP_URL}/inquiry/${id}`, { timeout: 3_000 });
      await page.waitForSelector('form', { timeout: 3_000 });

      responseMessage = 'Thank you for your restock request. We have restocked 10 units of the product.';
    } else if (isRestockRequest) {
      responseMessage = 'Thank you for your restock request. However, no product was specified.';
    }

    await sleep(2000); // Wait for a second to mimic human behavior

    await page.type('textarea[name="response"]', responseMessage);
    await Promise.all([
      page.click('button[type="submit"]'),
      page.waitForNavigation({ timeout: 3_000 }),
    ]);
    await page.close();
  } catch (e) {
    console.error(e);
  }

  await context.close();

  console.log(`end: ${id}`);
};

app.post('/visit', async (req, res) => {
    const { id, type, productId } = req.body;
    const uuidRegex = /^[0-9a-f]{8}$/i;
    if (!id || typeof id !== 'string' || !uuidRegex.test(id)) {
      return res.status(400).send('Invalid URL');
    }

    queue.push({
      id,
      type: type || 'general',
      productId: productId ? parseInt(productId) : null
    });

    processQueue();

    return res.send('Visiting the URL...');
});

process.on('unhandledRejection', (reason, promise) => {
    console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});

app.listen(port, () => {
    console.log(`Bot listening at http://localhost:${port}`);
});