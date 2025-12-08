import { JSDOM } from 'jsdom';
import { URL } from 'url';

// ============= CONFIGURATION - EDIT THESE =============
const START_URL = 'https://en.wikipedia.org/wiki/DevOps';
const MAX_PAGES = 10;
const DELAY_MS = 1000;

// Update these for your setup
const OGGOLE_API_URL = 'http://localhost:8080/api/batch-pages';  // Change to https://oggole.dk/api/batch-pages for production
const API_KEY = 'jiLU8V+1kztFB0lxwa06cKPqXLOIPkymc4EGvoC/qXY=';  // Must match your .env file
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

    try {
        const response = await fetch(cleanUrl.href, {
            headers: {
                'User-Agent': 'OggoleCrawler/1.0 (Educational purposes)'
            }
        });

        if (!response.ok) {
            console.error(`Failed to fetch ${cleanUrl.href}: ${response.statusText}`);
            return [];
        }

        const html = await response.text();
        const { window } = new JSDOM(html);
        const document = window.document;

        const title = document.querySelector('h1')?.textContent.trim() || '';
        const mainContent = Array.from(document.querySelectorAll('.mw-parser-output p'))
            .map(p => p.textContent.trim())
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

        const links = Array.from(document.querySelectorAll('a'))
            .map(link => link.getAttribute('href'))
            .filter(href => href && href.startsWith('/wiki/') && !href.includes(':'))
            .map(href => `https://en.wikipedia.org${href}`)
            .slice(0, 3);

        return [pageData, ...links];
    } catch (error) {
        console.error(`✗ Error crawling ${cleanUrl.href}:`, error.message);
        return [];
    }
}

async function sendToOggole(pages) {
    try {
        console.log(`\nSending ${pages.length} pages to Oggole...`);

        const response = await fetch(OGGOLE_API_URL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY
            },
            body: JSON.stringify({ pages })
        });

        if (!response.ok) {
            const text = await response.text();
            throw new Error(`API responded with ${response.status}: ${text}`);
        }

        const result = await response.json();
        console.log(`✓ Success! Inserted ${result.inserted}/${result.total} pages`);
        return true;
    } catch (error) {
        console.error('✗ Error:', error.message);
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
            const [pageData, ...newLinks] = await crawlPage(currentUrl);

            if (pageData && pageData.content) {
                allPages.push(pageData);
            }

            toVisit.push(...newLinks.filter(link => !visitedUrls.has(link)));

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
