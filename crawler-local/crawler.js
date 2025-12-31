import * as cheerio from 'cheerio';
import { URL } from 'url';

// ============= CONFIGURATION - EDIT THESE =============
const START_URL = 'https://en.wikipedia.org/wiki/DevOps';
const MAX_PAGES = 10;
const DELAY_MS = 1000;
const TIMEOUT_MS = 10000;  // 10 second timeout for all fetch requests

// Update these for your setup
const OGGOLE_API_URL = process.env.OGGOLE_API_URL || 'http://localhost:8080/api/batch-pages';
const API_KEY = process.env.CRAWLER_API_KEY;

// Validate API key is set
if (!API_KEY) {
    console.error('ERROR: CRAWLER_API_KEY environment variable is not set!');
    console.error('Please create a .env file in crawler-local/ with: CRAWLER_API_KEY=your_key_here');
    process.exit(1);
}
// ======================================================

const visitedUrls = new Set();

function delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function crawlPage(pageUrl) {
    const cleanUrl = new URL(pageUrl);
    cleanUrl.hash = '';

    if (visitedUrls.has(cleanUrl.href)) {
        return [];
    }

    visitedUrls.add(cleanUrl.href);

    // Create AbortController for timeout handling
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), TIMEOUT_MS);

    try {
        const response = await fetch(cleanUrl.href, {
            headers: {
                'User-Agent': 'OggoleCrawler/1.0 (Educational purposes)'
            },
            signal: controller.signal
        });

        clearTimeout(timeoutId); // Clear timeout on successful fetch

        if (!response.ok) {
            console.error(`Failed to fetch ${cleanUrl.href}: ${response.statusText}`);
            return [];
        }

        const html = await response.text();
        const $ = cheerio.load(html);

        const title = $('h1').text().trim() || '';
        const mainContent = $('.mw-parser-output p')
            .map((i, el) => $(el).text().trim())
            .get()
            .filter(text => text.length > 0)
            .join(' ')
            .replace(/[\t\n\r]+/g, ' ')
            .trim();

        const pageData = {
            title: title,
            url: cleanUrl.href,
            language: 'en',
            content: mainContent.substring(0, 250)  // Short preview for search results
        };

        console.log(`✓ Crawled: ${cleanUrl.href}`);

        const links = $('a')
            .map((i, el) => $(el).attr('href'))
            .get()
            .filter(href => href && href.startsWith('/wiki/') && !href.includes(':'))
            .map(href => `https://en.wikipedia.org${href}`)
            .slice(0, 10);  // Get more links to ensure we can crawl 10 pages

        return [pageData, ...links];
    } catch (error) {
        clearTimeout(timeoutId); // Clear timeout in error path

        // Handle timeout/abort errors specifically
        if (error.name === 'AbortError') {
            console.error(`✗ Timeout crawling ${cleanUrl.href}: Request took longer than ${TIMEOUT_MS / 1000}s`);
        } else {
            console.error(`✗ Error crawling ${cleanUrl.href}:`, error.message);
        }
        return [];
    }
}

async function sendToOggole(pages) {
    // Create AbortController for timeout handling
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), TIMEOUT_MS);

    try {
        console.log(`\nSending ${pages.length} pages to Oggole...`);

        const response = await fetch(OGGOLE_API_URL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY
            },
            body: JSON.stringify({ pages }),
            signal: controller.signal
        });

        clearTimeout(timeoutId); // Clear timeout on successful fetch

        if (!response.ok) {
            const text = await response.text();
            throw new Error(`API responded with ${response.status}: ${text}`);
        }

        const result = await response.json();
        console.log(`✓ Success! Inserted ${result.inserted}/${result.total} pages`);
        return true;
    } catch (error) {
        clearTimeout(timeoutId); // Clear timeout in error path

        // Handle timeout/abort errors specifically
        if (error.name === 'AbortError') {
            console.error(`✗ Timeout sending to Oggole: Request took longer than ${TIMEOUT_MS / 1000}s`);
        } else {
            console.error('✗ Error:', error.message);
        }
        return false;
    }
}

async function main() {
    console.log('=== Oggole Web Crawler ===');
    console.log(`Start URL: ${START_URL}`);
    console.log(`Max pages: ${MAX_PAGES}\n`);

    const allPages = [];
    const toVisit = [START_URL];

    while (toVisit.length > 0 && visitedUrls.size < MAX_PAGES) {
        const currentUrl = toVisit.shift();

        if (!visitedUrls.has(currentUrl)) {
            const result = await crawlPage(currentUrl);

            if (result.length > 0) {
                const pageData = result[0];
                const newLinks = result.slice(1);

                if (pageData && pageData.content) {
                    allPages.push(pageData);
                }

                toVisit.push(...newLinks.filter(link => !visitedUrls.has(link)));
            }

            if (toVisit.length > 0) {
                await delay(DELAY_MS);
            }
        }
    }

    console.log(`\nTotal pages crawled: ${allPages.length}`);

    if (allPages.length > 0) {
        await sendToOggole(allPages);
    }
}

main();
